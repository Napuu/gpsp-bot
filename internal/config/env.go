package config

import (
	"os"
	"reflect"
	"strings"
)

type Config struct {
	DISCORD_TOKEN     string
	TELEGRAM_TOKEN    string
	MISTRAL_TOKEN     string
	YTDLP_TMP_DIR     string
	EURIBOR_GRAPH_DIR string
	PROXY_URLS        string
	ENABLED_FEATURES  string
	EURIBOR_CSV_DIR   string
}

func FromEnv() Config {
	cfg := Config{
		YTDLP_TMP_DIR:     "/tmp/ytdlp",
		EURIBOR_GRAPH_DIR: "/tmp/euribor-graphs",
		EURIBOR_CSV_DIR:   "/tmp/euribor-exports",
	}
	v := reflect.ValueOf(&cfg).Elem()

	for i := range v.NumField() {
		field := v.Type().Field(i)
		envVar := field.Name
		envValue, exists := os.LookupEnv(envVar)
		if exists {
			v.Field(i).SetString(envValue)
		}
	}

	return cfg
}

func ProxyUrls() []string {
	return strings.Split(FromEnv().PROXY_URLS, ";")
}

func EnabledFeatures() []string {
	return strings.Split(FromEnv().ENABLED_FEATURES, ";")
}
