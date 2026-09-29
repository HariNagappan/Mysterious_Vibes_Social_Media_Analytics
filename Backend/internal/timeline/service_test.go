package timeline

import (
	"testing"
	"time"
)

func TestBucketMinutesFor(t *testing.T) {
	now := time.Now()
	cases := []struct {
		span time.Duration
		want int32
	}{
		{24 * time.Hour, 60},
		{5 * 24 * time.Hour, 720},
		{30 * 24 * time.Hour, 1440},
	}
	for _, tc := range cases {
		if got := bucketMinutesFor(now, now.Add(tc.span)); got != tc.want {
			t.Fatalf("span %s: expected %d, got %d", tc.span, tc.want, got)
		}
	}
}

func TestResolveWindowDefaults(t *testing.T) {
	from, to, err := resolveWindow(Query{TopicID: 1})
	if err != nil {
		t.Fatalf("resolve window: %v", err)
	}
	if !to.After(from) {
		t.Fatalf("expected from < to, got %s >= %s", from, to)
	}
	if to.Sub(from) != defaultWindow {
		t.Fatalf("expected default window %s, got %s", defaultWindow, to.Sub(from))
	}
}

func TestResolveWindowRejectsInverted(t *testing.T) {
	now := time.Now()
	_, _, err := resolveWindow(Query{TopicID: 1, From: now, HasFrom: true, To: now.Add(-time.Hour), HasTo: true})
	if err != ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}
