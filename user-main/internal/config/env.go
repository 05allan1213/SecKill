package config

import "os"

func ApplyEnvOverrides(c *Config) {
	overrideString(&c.Data.Database.Password, "USER_MYSQL_PASSWORD", "MYSQL_PASSWORD")
	overrideString(&c.Data.Redis.PassWord, "USER_REDIS_PASSWORD", "REDIS_PASSWORD")
}

func overrideString(target *string, keys ...string) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			*target = value
			return
		}
	}
}
