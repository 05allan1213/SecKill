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

	ApplyEnvOverrides(&cfg)

	if cfg.Data.Database.Password != "mysql-from-env" {
		t.Fatalf("expected mysql password override, got %q", cfg.Data.Database.Password)
	}
	if len(cfg.Data.Kafka.Producer.Brokers) != 2 || cfg.Data.Kafka.Producer.Brokers[0] != "127.0.0.1:9093" {
		t.Fatalf("unexpected producer brokers: %#v", cfg.Data.Kafka.Producer.Brokers)
	}
	if len(cfg.Data.Kafka.Consumer.Brokers) != 2 || cfg.Data.Kafka.Consumer.Brokers[1] != "127.0.0.1:9094" {
		t.Fatalf("unexpected consumer brokers: %#v", cfg.Data.Kafka.Consumer.Brokers)
	}
}
