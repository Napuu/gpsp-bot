package handlers

import (
	"testing"

	"github.com/napuu/gpsp-bot/internal/config"
)

// TestGenericMessageHandler_PingCommand tests that the GenericMessageHandler
// correctly parses the /ping command
func TestGenericMessageHandler_PingCommand(t *testing.T) {
	// Set up environment for testing
	t.Setenv("ENABLED_FEATURES", "ping")

	tests := []struct {
		name           string
		rawText        string
		expectedAction Action
	}{
		{
			name:           "ping with slash prefix",
			rawText:        "/ping",
			expectedAction: Ping,
		},
		{
			name:           "ping with exclamation prefix",
			rawText:        "!ping",
			expectedAction: Ping,
		},
		{
			name:           "ping with exclamation suffix",
			rawText:        "ping!",
			expectedAction: Ping,
		},
		{
			name:           "ping with additional text",
			rawText:        "/ping hello",
			expectedAction: Ping,
		},
		{
			name:           "not a ping command",
			rawText:        "hello world",
			expectedAction: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with the test input
			ctx := &Context{
				rawText: tt.rawText,
			}

			// Create a mock next handler to capture the result
			mockNext := &mockHandler{}
			handler := &GenericMessageHandler{next: mockNext}

			// Execute the handler
			handler.Execute(ctx)

			// Verify the action was set correctly
			if ctx.action != tt.expectedAction {
				t.Errorf("Expected action %q, got %q", tt.expectedAction, ctx.action)
			}
		})
	}
}

// TestConstructTextResponseHandler_Ping tests that the ConstructTextResponseHandler
// generates the correct "pong" response for ping commands
func TestConstructTextResponseHandler_Ping(t *testing.T) {
	ctx := &Context{
		action: Ping,
	}

	mockNext := &mockHandler{}
	handler := &ConstructTextResponseHandler{next: mockNext}

	handler.Execute(ctx)

	expectedResponse := "pong"
	if ctx.textResponse != expectedResponse {
		t.Errorf("Expected text response %q, got %q", expectedResponse, ctx.textResponse)
	}
}

// TestConstructTextResponseHandler_NoAction tests that the handler doesn't set
// a response when there's no action
func TestConstructTextResponseHandler_NoAction(t *testing.T) {
	ctx := &Context{
		action: "",
	}

	mockNext := &mockHandler{}
	handler := &ConstructTextResponseHandler{next: mockNext}

	handler.Execute(ctx)

	if ctx.textResponse != "" {
		t.Errorf("Expected empty text response, got %q", ctx.textResponse)
	}
}

// TestPingCommandEndToEnd tests the complete flow of processing a /ping command
func TestPingCommandEndToEnd(t *testing.T) {
	// Set up environment for testing
	t.Setenv("ENABLED_FEATURES", "ping")

	ctx := &Context{
		rawText: "/ping",
	}

	// Create the handler chain (simplified version for testing)
	genericMessageHandler := &GenericMessageHandler{}
	constructTextResponseHandler := &ConstructTextResponseHandler{}
	endHandler := &mockHandler{}

	genericMessageHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(endHandler)

	// Process the message
	genericMessageHandler.Execute(ctx)

	// Verify the action was set
	if ctx.action != Ping {
		t.Errorf("Expected action %q, got %q", Ping, ctx.action)
	}

	// Verify the response was set
	expectedResponse := "pong"
	if ctx.textResponse != expectedResponse {
		t.Errorf("Expected text response %q, got %q", expectedResponse, ctx.textResponse)
	}
}

// TestEnabledFeaturesValidation tests that commands are properly validated
func TestEnabledFeaturesValidation(t *testing.T) {
	tests := []struct {
		name          string
		rawText       string
		enabledFeats  string
		shouldSetAction bool
	}{
		{
			name:          "ping enabled and requested",
			rawText:       "/ping",
			enabledFeats:  "ping",
			shouldSetAction: true,
		},
		{
			name:          "ping not enabled",
			rawText:       "/ping",
			enabledFeats:  "dl",
			shouldSetAction: false,
		},
		{
			name:          "ping enabled with other features",
			rawText:       "/ping",
			enabledFeats:  "ping;dl;euribor",
			shouldSetAction: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment for this test
			t.Setenv("ENABLED_FEATURES", tt.enabledFeats)

			// Reload config from environment
			_ = config.FromEnv()

			ctx := &Context{
				rawText: tt.rawText,
			}

			mockNext := &mockHandler{}
			handler := &GenericMessageHandler{next: mockNext}

			handler.Execute(ctx)

			if tt.shouldSetAction && ctx.action != Ping {
				t.Errorf("Expected action to be set to %q, got %q", Ping, ctx.action)
			}
			if !tt.shouldSetAction && ctx.action != "" {
				t.Errorf("Expected action to be empty, got %q", ctx.action)
			}
		})
	}
}

// mockHandler is a simple mock implementation of ContextHandler for testing
type mockHandler struct {
	executed bool
}

func (m *mockHandler) Execute(ctx *Context) {
	m.executed = true
}

func (m *mockHandler) SetNext(next ContextHandler) {
	// Not needed for testing
}
