package handlers

import (
	"testing"
)

// TestPingCommand tests the complete flow of processing a /ping command
func TestPingCommand(t *testing.T) {
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
