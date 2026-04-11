package config

import "testing"

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
	t.Setenv("USER_MYSQL_PASSWORD", "mysql-from-env")
	t.Setenv("USER_REDIS_PASSWORD", "redis-from-env")
	t.Setenv("USER_TRACE_ENABLED", "true")
	t.Setenv("USER_TRACE_SAMPLER", "0.5")
	t.Setenv("USER_LOG_MAX_SIZE_MB", "16")
	t.Setenv("USER_LOG_MAX_BACKUPS", "4")
	t.Setenv("USER_LOG_COMPRESS", "false")
	t.Setenv("USER_ACCESS_SUMMARY_MAX_BYTES", "80")

	cfg := Config{
		Data: DataConf{
			Database: DatabaseConf{Password: "mysql-from-config"},
			Redis:    RedisConf{PassWord: "redis-from-config"},
		},
	}
	cfg.Telemetry.Sampler = 0.01

	ApplyEnvOverrides(&cfg)
	applyObservability(&cfg)

	if cfg.Data.Database.Password != "mysql-from-env" {
		t.Fatalf("expected mysql password override, got %q", cfg.Data.Database.Password)
	}
	if cfg.Data.Redis.PassWord != "redis-from-env" {
		t.Fatalf("expected redis password override, got %q", cfg.Data.Redis.PassWord)
	}
	if !cfg.Observability.Trace.Enabled {
		t.Fatal("expected trace to be enabled from env")
	}
	if cfg.Telemetry.Sampler != 0.5 {
		t.Fatalf("expected sampler override, got %v", cfg.Telemetry.Sampler)
	}
	if cfg.Observability.LogRotation.MaxSizeMB != 16 {
		t.Fatalf("expected max size override, got %d", cfg.Observability.LogRotation.MaxSizeMB)
	}
	if cfg.Observability.LogRotation.MaxBackups != 4 {
		t.Fatalf("expected max backups override, got %d", cfg.Observability.LogRotation.MaxBackups)
	}
	if cfg.Observability.LogRotation.Compress {
		t.Fatal("expected compress override to be false")
	}
	if cfg.Observability.AccessLog.SummaryMaxBytes != 80 {
		t.Fatalf("expected access summary bytes override, got %d", cfg.Observability.AccessLog.SummaryMaxBytes)
	}
}

func TestApplyObservability_DisablesTrace(t *testing.T) {
	cfg := Config{}
	cfg.Telemetry.Name = "user"
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
