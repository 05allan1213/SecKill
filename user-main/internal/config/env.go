package config

import (
	"os"
	"strconv"
)

func ApplyEnvOverrides(c *Config) {
	overrideString(&c.Data.Database.Password, "USER_MYSQL_PASSWORD", "MYSQL_PASSWORD")
	overrideString(&c.Data.Redis.PassWord, "USER_REDIS_PASSWORD", "REDIS_PASSWORD")
	overrideBool(&c.Observability.Trace.Enabled, "USER_TRACE_ENABLED", "TRACE_ENABLED")
	overrideFloat64(&c.Telemetry.Sampler, "USER_TRACE_SAMPLER", "TRACE_SAMPLER")
	overrideInt(&c.Observability.LogRotation.MaxSizeMB, "USER_LOG_MAX_SIZE_MB", "LOG_MAX_SIZE_MB")
	overrideInt(&c.Observability.LogRotation.MaxBackups, "USER_LOG_MAX_BACKUPS", "LOG_MAX_BACKUPS")
	overrideBool(&c.Observability.LogRotation.Compress, "USER_LOG_COMPRESS", "LOG_COMPRESS")
	overrideInt(&c.Observability.AccessLog.SummaryMaxBytes, "USER_ACCESS_SUMMARY_MAX_BYTES", "ACCESS_SUMMARY_MAX_BYTES")
}

func overrideString(target *string, keys ...string) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			*target = value
			return
		}
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
