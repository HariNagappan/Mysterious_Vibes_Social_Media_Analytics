package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Handler processes a single event. Returning an error triggers the retry /
// dead-letter policy of the consumer.
type Handler func(ctx context.Context, evt Event) error

// Consumer reads events from one stream using a Redis consumer group. The same
// structure maps onto a Kafka consumer group later (group id, single active
// consumer per partition, manual ack).
type Consumer struct {
	client     *redis.Client
	stream     string
	group      string
	name       string
	handler    Handler
	logger     *zap.Logger
	batch      int64
	block      time.Duration
	maxRetries int64
}

// NewConsumer builds a consumer for one stream.
func NewConsumer(client *redis.Client, stream, group, name string, handler Handler, logger *zap.Logger, batch int64, block time.Duration, maxRetries int64) *Consumer {
	if batch <= 0 {
		batch = 16
	}
	if block <= 0 {
		block = 2 * time.Second
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &Consumer{
		client:     client,
		stream:     stream,
		group:      group,
		name:       name,
		handler:    handler,
		logger:     logger,
		batch:      batch,
		block:      block,
		maxRetries: maxRetries,
	}
}

// EnsureGroup creates the consumer group (idempotent).
func (c *Consumer) EnsureGroup(ctx context.Context) error {
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("pipeline: create group for %s: %w", c.stream, err)
	}
	return nil
}

// Run blocks until ctx is cancelled, reading and processing events.
func (c *Consumer) Run(ctx context.Context) error {
	if err := c.EnsureGroup(ctx); err != nil {
		return err
	}
	c.reclaimStale(ctx)

	for {
		if ctx.Err() != nil {
			return nil
		}
		res, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.name,
			Streams:  []string{c.stream, ">"},
			Count:    c.batch,
			Block:    c.block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Warn("pipeline: read failed", zap.String("stream", c.stream), zap.Error(err))
			if !sleepCtx(ctx, time.Second) {
				return nil
			}
			continue
		}
		for _, stream := range res {
			for _, msg := range stream.Messages {
				c.processMessage(ctx, msg)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg redis.XMessage) {
	evt, err := eventFromMessage(msg)
	if err != nil {
		c.logger.Error("pipeline: dropping malformed event",
			zap.String("stream", c.stream), zap.String("message_id", msg.ID), zap.Error(err))
		_ = c.client.XAck(ctx, c.stream, c.group, msg.ID).Err()
		return
	}

	if err := c.handler(ctx, evt); err != nil {
		retries := c.retryCount(ctx, msg.ID)
		if retries >= c.maxRetries {
			c.deadLetter(ctx, msg, evt, err)
			_ = c.client.XAck(ctx, c.stream, c.group, msg.ID).Err()
			return
		}
		c.logger.Warn("pipeline: handler failed; leaving message pending for retry",
			zap.String("stream", c.stream),
			zap.String("event", string(evt.Type)),
			zap.Int64("retry_count", retries),
			zap.Error(err))
		return
	}

	if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
		c.logger.Warn("pipeline: ack failed", zap.String("stream", c.stream), zap.String("message_id", msg.ID), zap.Error(err))
	}
}

// retryCount returns how often a message has already been delivered.
func (c *Consumer) retryCount(ctx context.Context, id string) int64 {
	entries, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  id,
		End:    id,
		Count:  1,
	}).Result()
	if err != nil || len(entries) == 0 {
		return 0
	}
	return entries[0].RetryCount
}

// deadLetter copies a permanently failing message to `<stream>:dead`.
func (c *Consumer) deadLetter(ctx context.Context, msg redis.XMessage, evt Event, handlerErr error) {
	values := map[string]any{
		"original_id": msg.ID,
		"type":        string(evt.Type),
		"payload":     string(evt.Payload),
		"error":       handlerErr.Error(),
		"failed_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	if _, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: c.stream + ":dead",
		MaxLen: 5000,
		Approx: true,
		Values: values,
	}).Result(); err != nil {
		c.logger.Error("pipeline: dead-letter write failed", zap.Error(err))
	}
	c.logger.Error("pipeline: event moved to dead-letter stream",
		zap.String("stream", c.stream),
		zap.String("event", string(evt.Type)),
		zap.Error(handlerErr))
}

// reclaimStale re-delivers messages left pending by previous crashed
// consumers (older than 5 minutes) so processing eventually completes.
func (c *Consumer) reclaimStale(ctx context.Context) {
	start := "0-0"
	for i := 0; i < 10; i++ {
		msgs, next, err := c.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   c.stream,
			Group:    c.group,
			Consumer: c.name,
			MinIdle:  5 * time.Minute,
			Start:    start,
			Count:    32,
		}).Result()
		if err != nil {
			c.logger.Debug("pipeline: auto-claim skipped", zap.String("stream", c.stream), zap.Error(err))
			return
		}
		for _, msg := range msgs {
			c.processMessage(ctx, msg)
		}
		if next == "" || next == "0-0" || len(msgs) == 0 {
			return
		}
		start = next
	}
}

func eventFromMessage(msg redis.XMessage) (Event, error) {
	get := func(key string) string {
		if v, ok := msg.Values[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", v)
		}
		return ""
	}

	id := get("id")
	evtType := get("type")
	payload := get("payload")
	if evtType == "" || payload == "" {
		return Event{}, fmt.Errorf("pipeline: event missing required fields (message %s)", msg.ID)
	}
	occurredAt := time.Now().UTC()
	if ts := get("occurred_at"); ts != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			occurredAt = parsed
		}
	}
	return Event{
		ID:         id,
		Type:       EventType(evtType),
		Version:    1,
		OccurredAt: occurredAt,
		Payload:    []byte(payload),
	}, nil
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
