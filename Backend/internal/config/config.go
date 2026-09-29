// Package config loads application configuration from environment variables.
//
// Every deployment target (local development, docker-compose, Kubernetes)
// supplies the same set of variables, documented in .env.example. A .env file
// in the process working directory is loaded automatically for local
// development; real environment variables always win over it.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the root configuration object shared by the api, worker and
// scheduler binaries.
type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	JWT      JWTConfig
	ML       MLConfig
	Search   SearchConfig
	Pipeline PipelineConfig
	Demo     DemoConfig
}

// AppConfig holds process-level settings.
type AppConfig struct {
	Env           string
	Port          int
	CORSOrigins   []string
	RunMigrations bool
}

// PostgresConfig holds the PostgreSQL connection settings.
type PostgresConfig struct {
	URL      string
	MaxConns int32
}

// RedisConfig holds Redis (cache + Redis Streams) settings.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	CacheTTL time.Duration
}

// JWTConfig holds token signing settings.
type JWTConfig struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

// MLConfig points at the Python FastAPI microservices.
type MLConfig struct {
	SentimentURL  string
	EmbeddingURL  string
	SummarizerURL string
	Timeout       time.Duration
}

// SearchConfig configures the Elasticsearch/OpenSearch abstraction.
type SearchConfig struct {
	Enabled  bool
	URL      string
	Index    string
	Username string
	Password string
}

// PipelineConfig configures the Redis Streams event pipeline.
type PipelineConfig struct {
	StreamPrefix  string
	ConsumerGroup string
	BatchSize     int64
	Block         time.Duration
	MaxRetries    int
}

// DemoConfig configures the built-in demo ingestion generator and the
// scheduler cadence.
type DemoConfig struct {
	IngestEnabled  bool
	IngestInterval time.Duration
	ScheduleEvery  time.Duration
}

// Load reads configuration, applying development-friendly defaults so the
// stack can boot with a bare `docker compose up`.
func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		App: AppConfig{
			Env:           getEnv("APP_ENV", "development"),
			Port:          getEnvInt("APP_PORT", 8080),
			CORSOrigins:   getEnvList("CORS_ORIGINS", []string{"*"}),
			RunMigrations: getEnvBool("RUN_MIGRATIONS", true),
		},
		Postgres: PostgresConfig{
			URL:      getEnv("DATABASE_URL", "postgres://pulsegraph:pulsegraph@localhost:5432/pulsegraph?sslmode=disable"),
			MaxConns: int32(getEnvInt("POSTGRES_MAX_CONNS", 10)),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			CacheTTL: getEnvDuration("CACHE_TTL", 60*time.Second),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			TTL:    getEnvDuration("JWT_TTL", 24*time.Hour),
			Issuer: getEnv("JWT_ISSUER", "pulsegraph"),
		},
		ML: MLConfig{
			SentimentURL:  getEnv("ML_SENTIMENT_URL", "http://localhost:8001"),
			EmbeddingURL:  getEnv("ML_EMBEDDING_URL", "http://localhost:8002"),
			SummarizerURL: getEnv("ML_SUMMARIZER_URL", "http://localhost:8003"),
			Timeout:       getEnvDuration("ML_TIMEOUT", 15*time.Second),
		},
		Search: SearchConfig{
			Enabled:  getEnvBool("SEARCH_ENABLED", false),
			URL:      getEnv("SEARCH_URL", "http://localhost:9200"),
			Index:    getEnv("SEARCH_INDEX", "pulsegraph-posts"),
			Username: getEnv("SEARCH_USERNAME", ""),
			Password: getEnv("SEARCH_PASSWORD", ""),
		},
		Pipeline: PipelineConfig{
			StreamPrefix:  getEnv("PIPELINE_STREAM_PREFIX", "pulsegraph"),
			ConsumerGroup: getEnv("PIPELINE_CONSUMER_GROUP", "pulsegraph-workers"),
			BatchSize:     int64(getEnvInt("PIPELINE_BATCH_SIZE", 16)),
			Block:         getEnvDuration("PIPELINE_BLOCK", 2*time.Second),
			MaxRetries:    getEnvInt("PIPELINE_MAX_RETRIES", 3),
		},
		Demo: DemoConfig{
			IngestEnabled:  getEnvBool("DEMO_INGEST_ENABLED", false),
			IngestInterval: getEnvDuration("DEMO_INGEST_INTERVAL", 30*time.Second),
			ScheduleEvery:  getEnvDuration("SCHEDULE_INTERVAL", 2*time.Minute),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("config: APP_PORT %d is out of range", c.App.Port)
	}
	if c.Postgres.URL == "" {
		return fmt.Errorf("config: DATABASE_URL must be set")
	}
	if c.JWT.Secret == "" {
		if c.App.Env == "production" {
			return fmt.Errorf("config: JWT_SECRET is required in production")
		}
		c.JWT.Secret = "pulsegraph-development-secret"
	}
	if c.ML.Timeout <= 0 {
		c.ML.Timeout = 15 * time.Second
	}
	return nil
}

// loadDotEnv reads simple KEY=VALUE lines from path and sets variables that
// are not already present in the process environment. A missing file is fine.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	switch strings.ToLower(getEnv(key, "")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if seconds, err := strconv.Atoi(v); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
