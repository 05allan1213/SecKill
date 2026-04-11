package config

import (
	"os"
	"strconv"
	"strings"
)

func ApplyEnvOverrides(c *Config) {
	overrideString(&c.Data.Database.Password, "SECKILL_MYSQL_PASSWORD", "MYSQL_PASSWORD")
	overrideString(&c.Data.Redis.PassWord, "SECKILL_REDIS_PASSWORD", "REDIS_PASSWORD")
	overrideCSV(&c.Data.Kafka.Producer.Brokers, "SECKILL_KAFKA_PRODUCER_BROKERS", "SECKILL_KAFKA_BROKERS", "KAFKA_BROKERS")
	overrideCSV(&c.Data.Kafka.Consumer.Brokers, "SECKILL_KAFKA_CONSUMER_BROKERS", "SECKILL_KAFKA_BROKERS", "KAFKA_BROKERS")
	overrideBool(&c.Observability.Trace.Enabled, "SECKILL_TRACE_ENABLED", "TRACE_ENABLED")
	overrideFloat64(&c.Telemetry.Sampler, "SECKILL_TRACE_SAMPLER", "TRACE_SAMPLER")
	overrideInt(&c.Observability.LogRotation.MaxSizeMB, "SECKILL_LOG_MAX_SIZE_MB", "LOG_MAX_SIZE_MB")
	overrideInt(&c.Observability.LogRotation.MaxBackups, "SECKILL_LOG_MAX_BACKUPS", "LOG_MAX_BACKUPS")
	overrideBool(&c.Observability.LogRotation.Compress, "SECKILL_LOG_COMPRESS", "LOG_COMPRESS")
	overrideInt(&c.Observability.AccessLog.SummaryMaxBytes, "SECKILL_ACCESS_SUMMARY_MAX_BYTES", "ACCESS_SUMMARY_MAX_BYTES")
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

func overrideInt(target *int, keys ...string) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err == nil {
			*target = parsed
		}
		return
	}
}

func overrideBool(target *bool, keys ...string) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			*target = parsed
		}
		return
	}
}

func overrideFloat64(target *float64, keys ...string) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			*target = parsed
		}
		return
	}
}
