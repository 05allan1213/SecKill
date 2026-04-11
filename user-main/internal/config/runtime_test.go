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

	cfg := Config{
		Data: DataConf{
			Database: DatabaseConf{Password: "mysql-from-config"},
			Redis:    RedisConf{PassWord: "redis-from-config"},
		},
	}

	ApplyEnvOverrides(&cfg)

	if cfg.Data.Database.Password != "mysql-from-env" {
		t.Fatalf("expected mysql password override, got %q", cfg.Data.Database.Password)
	}
	if cfg.Data.Redis.PassWord != "redis-from-env" {
		t.Fatalf("expected redis password override, got %q", cfg.Data.Redis.PassWord)
	}
}
