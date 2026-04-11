package config

import "os"

func ApplyEnvOverrides(c *Config) {
	overrideString(&c.Auth.Secret, "GATEWAY_AUTH_SECRET", "AUTH_SECRET")
	overrideString(&c.Redis.PassWord, "GATEWAY_REDIS_PASSWORD", "REDIS_PASSWORD")
}

func overrideString(target *string, keys ...string) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			*target = value
			return
		}
	}
}
