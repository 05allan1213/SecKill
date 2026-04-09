package config

import (
	"os"
	"strings"
)

func ApplyEnvOverrides(c *Config) {
	overrideString(&c.Data.Database.Password, "SECKILL_MYSQL_PASSWORD", "MYSQL_PASSWORD")
	overrideString(&c.Data.Redis.PassWord, "SECKILL_REDIS_PASSWORD", "REDIS_PASSWORD")
	overrideCSV(&c.Data.Kafka.Producer.Brokers, "SECKILL_KAFKA_PRODUCER_BROKERS", "SECKILL_KAFKA_BROKERS", "KAFKA_BROKERS")
	overrideCSV(&c.Data.Kafka.Consumer.Brokers, "SECKILL_KAFKA_CONSUMER_BROKERS", "SECKILL_KAFKA_BROKERS", "KAFKA_BROKERS")
}

func overrideString(target *string, keys ...string) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			*target = value
			return
		}
	}
}

func overrideCSV(target *[]string, keys ...string) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		parts := strings.Split(value, ",")
		normalized := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				normalized = append(normalized, part)
			}
		}
		if len(normalized) > 0 {
			*target = normalized
		}
		return
	}
}
