package config

import (
	"os"
	"testing"
)

func TestApplyRuntimeConfig(t *testing.T) {
	cfg := Config{
		Data: DataConf{
			Database: DatabaseConf{Addr: "db-old"},
			Redis:    RedisConf{Addr: "redis-old"},
		},
	}

	ApplyRuntimeConfig(&cfg, RuntimeConfig{
		Data: &DataConf{
			Database: DatabaseConf{Addr: "db-new"},
			Redis:    RedisConf{Addr: "redis-new"},
		},
	})

	if cfg.Data.Database.Addr != "db-new" {
		t.Fatalf("expected database addr override, got %q", cfg.Data.Database.Addr)
	}
	if cfg.Data.Redis.Addr != "redis-new" {
		t.Fatalf("expected redis addr override, got %q", cfg.Data.Redis.Addr)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("SECKILL_MYSQL_PASSWORD", "mysql-from-env")
	t.Setenv("SECKILL_KAFKA_BROKERS", "127.0.0.1:9093,127.0.0.1:9094")
	t.Setenv("SECKILL_TRACE_ENABLED", "true")
	t.Setenv("SECKILL_TRACE_SAMPLER", "0.2")
	t.Setenv("SECKILL_LOG_MAX_SIZE_MB", "24")
	t.Setenv("SECKILL_LOG_MAX_BACKUPS", "6")
	t.Setenv("SECKILL_LOG_COMPRESS", "false")
	t.Setenv("SECKILL_ACCESS_SUMMARY_MAX_BYTES", "72")
	_ = os.Unsetenv("MYSQL_PASSWORD")
	_ = os.Unsetenv("KAFKA_BROKERS")

	cfg := Config{
		Data: DataConf{
			Database: DatabaseConf{Password: "mysql-from-config"},
			Kafka: KafkaConf{
				Producer: KafkaProducerConf{Brokers: []string{"old"}},
				Consumer: KafkaConsumerConf{Brokers: []string{"old"}},
			},
		},
	}
	cfg.Telemetry.Sampler = 0.01

	ApplyEnvOverrides(&cfg)
	applyObservability(&cfg)

	if cfg.Data.Database.Password != "mysql-from-env" {
		t.Fatalf("expected mysql password override, got %q", cfg.Data.Database.Password)
	}
	if len(cfg.Data.Kafka.Producer.Brokers) != 2 || cfg.Data.Kafka.Producer.Brokers[0] != "127.0.0.1:9093" {
		t.Fatalf("unexpected producer brokers: %#v", cfg.Data.Kafka.Producer.Brokers)
	}
	if len(cfg.Data.Kafka.Consumer.Brokers) != 2 || cfg.Data.Kafka.Consumer.Brokers[1] != "127.0.0.1:9094" {
		t.Fatalf("unexpected consumer brokers: %#v", cfg.Data.Kafka.Consumer.Brokers)
	}
	if !cfg.Observability.Trace.Enabled {
		t.Fatal("expected trace to be enabled from env")
	}
	if cfg.Telemetry.Sampler != 0.2 {
		t.Fatalf("expected sampler override, got %v", cfg.Telemetry.Sampler)
	}
	if cfg.Observability.LogRotation.MaxSizeMB != 24 {
		t.Fatalf("expected max size override, got %d", cfg.Observability.LogRotation.MaxSizeMB)
	}
	if cfg.Observability.LogRotation.MaxBackups != 6 {
		t.Fatalf("expected max backups override, got %d", cfg.Observability.LogRotation.MaxBackups)
	}
	if cfg.Observability.LogRotation.Compress {
		t.Fatal("expected compress override to be false")
	}
	if cfg.Observability.AccessLog.SummaryMaxBytes != 72 {
		t.Fatalf("expected access summary bytes override, got %d", cfg.Observability.AccessLog.SummaryMaxBytes)
	}
}

func TestApplyObservability_DisablesTrace(t *testing.T) {
	cfg := Config{}
	cfg.Telemetry.Name = "seckill"
	cfg.Telemetry.Endpoint = "logs/trace.json"
	cfg.Telemetry.Batcher = "file"
	cfg.Telemetry.Sampler = 1
	cfg.Middlewares.Trace = true

	applyObservability(&cfg)

	if cfg.Middlewares.Trace {
		t.Fatal("expected trace middleware to be disabled")
	}
	if cfg.Telemetry.Endpoint != "" || cfg.Telemetry.Name != "" || cfg.Telemetry.Batcher != "" || cfg.Telemetry.Sampler != 0 {
		t.Fatalf("expected telemetry to be disabled, got %#v", cfg.Telemetry)
	}
}
