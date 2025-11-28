package version

import (
	"testing"
)

// TestDefaultVersion verifies that when no ldflags are used,
// the Version variable defaults to "dev".
// When testing with ldflags (e.g., go test -ldflags "-X github.com/napuu/gpsp-bot/internal/version.Version=123" ./internal/version/...),
// this test will fail as expected, confirming ldflags injection works.
func TestDefaultVersion(t *testing.T) {
	expectedVersion := "dev"
	if Version != expectedVersion {
		t.Errorf("Expected default version %q, got %q", expectedVersion, Version)
	}
}

// TestVersionIsNotEmpty verifies that the Version variable is never empty.
// This test passes both with and without ldflags.
func TestVersionIsNotEmpty(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
