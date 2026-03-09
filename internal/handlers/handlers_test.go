package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/napuu/gpsp-bot/internal/version"
	"github.com/napuu/gpsp-bot/pkg/utils"
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

// TestVersionCommand tests the complete flow of processing a /version command
func TestVersionCommand(t *testing.T) {
	// Set up environment for testing
	t.Setenv("ENABLED_FEATURES", "version")

	ctx := &Context{
		rawText: "/version",
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
	if ctx.action != Version {
		t.Errorf("Expected action %q, got %q", Version, ctx.action)
	}

	// Verify the response was set to the human-readable version
	expectedResponse := version.GetHumanReadableVersion()
	if ctx.textResponse != expectedResponse {
		t.Errorf("Expected text response %q, got %q", expectedResponse, ctx.textResponse)
	}
}

// TestEuriborTextResponseIsSetBeforeImageHandler verifies that the text response
// is constructed before the image handler runs, so the caption is included with the graph.
func TestEuriborTextResponseIsSetBeforeImageHandler(t *testing.T) {
	rateDate, err := time.Parse("2006-01-02", "2024-03-01")
	if err != nil {
		t.Fatalf("failed to parse date: %v", err)
	}
	ctx := &Context{
		action: Euribor,
		rates: utils.EuriborRateEntry{
			Date:         rateDate,
			ThreeMonths:  3.852,
			SixMonths:    3.921,
			TwelveMonths: 3.732,
		},
		finalImagePath: "/tmp/some-graph.jpg",
	}

	constructTextResponseHandler := &ConstructTextResponseHandler{}
	captureHandler := &mockHandler{}
	constructTextResponseHandler.SetNext(captureHandler)

	constructTextResponseHandler.Execute(ctx)

	if ctx.textResponse == "" {
		t.Error("Expected textResponse to be set for Euribor action, but it was empty")
	}
	if !strings.Contains(ctx.textResponse, "01.03.") {
		t.Errorf("Expected textResponse to contain the date, got: %q", ctx.textResponse)
	}
	if !strings.Contains(ctx.textResponse, "3.732") {
		t.Errorf("Expected textResponse to contain 12-month rate, got: %q", ctx.textResponse)
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
