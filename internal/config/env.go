package config

import (
	"os"
	"reflect"
	"slices"
	"strings"
)

type Config struct {
	DISCORD_TOKEN      string
	TELEGRAM_TOKEN     string
	MISTRAL_TOKEN      string
	YTDLP_TMP_DIR      string
	EURIBOR_GRAPH_DIR  string
	PROXY_URLS         string
	ENABLED_FEATURES   string
	EURIBOR_CSV_DIR    string
	REPOST_DB_DIR      string
	TELETEXT_IMAGE_DIR string
	ALWAYS_RE_ENCODE   bool
}

func FromEnv() Config {
	cfg := Config{
		YTDLP_TMP_DIR:      "/tmp/ytdlp",
		EURIBOR_GRAPH_DIR:  "/tmp/euribor-graphs",
		EURIBOR_CSV_DIR:    "/tmp/euribor-exports",
		REPOST_DB_DIR:      "/tmp/repost-db",
		TELETEXT_IMAGE_DIR: "/tmp/teletext",
		ALWAYS_RE_ENCODE:   false,
	}
	v := reflect.ValueOf(&cfg).Elem()

	for i := range v.NumField() {
		field := v.Type().Field(i)
		envVar := field.Name
		envValue, exists := os.LookupEnv(envVar)
		if exists {
			if field.Type == reflect.TypeOf(cfg.ALWAYS_RE_ENCODE) {
				truthyValues := []string{"true", "yes", "1"}
				valueToLower := strings.ToLower(envValue)
				isTruthy := slices.Contains(truthyValues, valueToLower)
				v.Field(i).SetBool(isTruthy)
			} else {
				v.Field(i).SetString(envValue)
			}
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

// backgroundFeatures are ENABLED_FEATURES values that run in the background and are not user commands.
var backgroundFeatures = map[string]struct{}{
	"daymeme": {},
}

// IsBackgroundFeature reports whether a feature flag is background-only (not a slash command).
func IsBackgroundFeature(feature string) bool {
	_, ok := backgroundFeatures[feature]
	return ok
}
