package platforms

import (
	"testing"

	"github.com/napuu/gpsp-bot/internal/handlers"
)

// TestActionExists tests that the actionExists function correctly identifies valid actions
func TestActionExists(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{
			name:     "ping action exists",
			action:   "ping",
			expected: true,
		},
		{
			name:     "dl action exists",
			action:   "dl",
			expected: true,
		},
		{
			name:     "euribor action exists",
			action:   "euribor",
			expected: true,
		},
		{
			name:     "tuplilla action exists",
			action:   "tuplilla",
			expected: true,
		},
		{
			name:     "non-existent action",
			action:   "nonexistent",
			expected: false,
		},
		{
			name:     "empty action",
			action:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := actionExists(tt.action)
			if result != tt.expected {
				t.Errorf("actionExists(%q) = %v, want %v", tt.action, result, tt.expected)
			}
		})
	}
}

// TestVerifyEnabledCommands tests that VerifyEnabledCommands panics for invalid commands
func TestVerifyEnabledCommands(t *testing.T) {
	tests := []struct {
		name        string
		enabledFeats string
		shouldPanic bool
	}{
		{
			name:        "valid single command",
			enabledFeats: "ping",
			shouldPanic: false,
		},
		{
			name:        "valid multiple commands",
			enabledFeats: "ping;dl;euribor",
			shouldPanic: false,
		},
		{
			name:        "invalid command",
			enabledFeats: "invalidcommand",
			shouldPanic: true,
		},
		{
			name:        "mix of valid and invalid commands",
			enabledFeats: "ping;invalidcommand",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENABLED_FEATURES", tt.enabledFeats)

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("VerifyEnabledCommands() did not panic for invalid command")
					}
				}()
			}

			VerifyEnabledCommands()
		})
	}
}

// TestActionMap verifies that all expected actions are present in the ActionMap
func TestActionMap(t *testing.T) {
	expectedActions := []handlers.Action{
		handlers.Ping,
		handlers.DownloadVideo,
		handlers.Euribor,
		handlers.Tuplilla,
	}

	for _, action := range expectedActions {
		t.Run(string(action), func(t *testing.T) {
			description, exists := handlers.ActionMap[action]
			if !exists {
				t.Errorf("Action %q not found in ActionMap", action)
			}
			if description == "" {
				t.Errorf("Action %q has empty description", action)
			}
		})
	}
}

// TestActionDescriptions verifies that all actions have meaningful descriptions
func TestActionDescriptions(t *testing.T) {
	expectedDescriptions := map[handlers.Action]string{
		handlers.Ping:          "Ping",
		handlers.DownloadVideo: "Lataa video",
		handlers.Euribor:       "Tuoreet Euribor-korot",
		handlers.Tuplilla:      "Tuplilla...",
	}

	for action, expectedDesc := range expectedDescriptions {
		t.Run(string(action), func(t *testing.T) {
			actualDesc, exists := handlers.ActionMap[action]
			if !exists {
				t.Errorf("Action %q not found in ActionMap", action)
				return
			}
			if string(actualDesc) != expectedDesc {
				t.Errorf("Action %q description = %q, want %q", action, actualDesc, expectedDesc)
			}
		})
	}
}
