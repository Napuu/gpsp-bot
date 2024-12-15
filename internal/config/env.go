package config

import (
	"os"
	"reflect"
)

type Config struct {
	DISCORD_TOKEN  string
	TELEGRAM_TOKEN string
	OPENAI_TOKEN   string
	YTDLP_TMP_DIR  string
	DATABASE_FILE  string
	PROXY_URLS     string
}

func FromEnv() Config {
	cfg := Config{
		YTDLP_TMP_DIR: "/tmp",
		DATABASE_FILE: "/tmp/cache.db",
	}
	v := reflect.ValueOf(&cfg).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		envVar := field.Name
		envValue, exists := os.LookupEnv(envVar)
		if exists {
			v.Field(i).SetString(envValue)
		}
	}

	return cfg
}
