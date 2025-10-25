package config

import (
	"testing"
)

// TestFromEnv_EnabledFeatures tests that ENABLED_FEATURES are correctly parsed
func TestFromEnv_EnabledFeatures(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "single feature",
			envValue: "ping",
			expected: "ping",
		},
		{
			name:     "multiple features",
			envValue: "ping;dl;euribor",
			expected: "ping;dl;euribor",
		},
		{
			name:     "empty features",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENABLED_FEATURES", tt.envValue)

			cfg := FromEnv()

			if cfg.ENABLED_FEATURES != tt.expected {
				t.Errorf("Expected ENABLED_FEATURES %q, got %q", tt.expected, cfg.ENABLED_FEATURES)
			}
		})
	}
}

// TestEnabledFeatures tests the EnabledFeatures helper function
func TestEnabledFeatures(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedLen   int
		expectedFirst string
	}{
		{
			name:          "single feature",
			envValue:      "ping",
			expectedLen:   1,
			expectedFirst: "ping",
		},
		{
			name:          "multiple features",
			envValue:      "ping;dl;euribor",
			expectedLen:   3,
			expectedFirst: "ping",
		},
		{
			name:          "empty features",
			envValue:      "",
			expectedLen:   1,
			expectedFirst: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENABLED_FEATURES", tt.envValue)

			features := EnabledFeatures()

			if len(features) != tt.expectedLen {
				t.Errorf("Expected %d features, got %d", tt.expectedLen, len(features))
			}

			if len(features) > 0 && features[0] != tt.expectedFirst {
				t.Errorf("Expected first feature to be %q, got %q", tt.expectedFirst, features[0])
			}
		})
	}
}

// TestFromEnv_Defaults tests that default values are set correctly
func TestFromEnv_Defaults(t *testing.T) {
	// Clear any environment variables that might interfere
	cfg := FromEnv()

	expectedDefaults := map[string]string{
		"YTDLP_TMP_DIR":     "/tmp/ytdlp",
		"EURIBOR_GRAPH_DIR": "/tmp/euribor-graphs",
		"EURIBOR_CSV_DIR":   "/tmp/euribor-exports",
	}

	if cfg.YTDLP_TMP_DIR != expectedDefaults["YTDLP_TMP_DIR"] {
		t.Errorf("Expected YTDLP_TMP_DIR default %q, got %q", expectedDefaults["YTDLP_TMP_DIR"], cfg.YTDLP_TMP_DIR)
	}

	if cfg.EURIBOR_GRAPH_DIR != expectedDefaults["EURIBOR_GRAPH_DIR"] {
		t.Errorf("Expected EURIBOR_GRAPH_DIR default %q, got %q", expectedDefaults["EURIBOR_GRAPH_DIR"], cfg.EURIBOR_GRAPH_DIR)
	}

	if cfg.EURIBOR_CSV_DIR != expectedDefaults["EURIBOR_CSV_DIR"] {
		t.Errorf("Expected EURIBOR_CSV_DIR default %q, got %q", expectedDefaults["EURIBOR_CSV_DIR"], cfg.EURIBOR_CSV_DIR)
	}

	if cfg.ALWAYS_RE_ENCODE != false {
		t.Errorf("Expected ALWAYS_RE_ENCODE default to be false, got %v", cfg.ALWAYS_RE_ENCODE)
	}
}

// TestFromEnv_AlwaysReEncode tests the ALWAYS_RE_ENCODE boolean parsing
func TestFromEnv_AlwaysReEncode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "yes",
			envValue: "yes",
			expected: true,
		},
		{
			name:     "1",
			envValue: "1",
			expected: true,
		},
		{
			name:     "TRUE (uppercase)",
			envValue: "TRUE",
			expected: true,
		},
		{
			name:     "false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "no",
			envValue: "no",
			expected: false,
		},
		{
			name:     "0",
			envValue: "0",
			expected: false,
		},
		{
			name:     "random string",
			envValue: "random",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ALWAYS_RE_ENCODE", tt.envValue)

			cfg := FromEnv()

			if cfg.ALWAYS_RE_ENCODE != tt.expected {
				t.Errorf("Expected ALWAYS_RE_ENCODE %v, got %v", tt.expected, cfg.ALWAYS_RE_ENCODE)
			}
		})
	}
}
