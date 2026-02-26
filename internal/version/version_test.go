package version

import (
	"strings"
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

// TestGetHumanReadableVersionDev tests the dev version output
func TestGetHumanReadableVersionDev(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "dev"
	result := GetHumanReadableVersion()
	expected := "gpsp-bot dev (development build)"
	if !strings.HasPrefix(result, expected) {
		t.Errorf("Expected result to start with %q, got %q", expected, result)
	}
}

// TestGetHumanReadableVersionTimestamp tests timestamp version parsing
func TestGetHumanReadableVersionTimestamp(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "20231128151500"
	result := GetHumanReadableVersion()
	expected := "gpsp-bot 20231128151500 (built 2023-11-28 15:15:00 UTC)"
	if !strings.HasPrefix(result, expected) {
		t.Errorf("Expected result to start with %q, got %q", expected, result)
	}
}

// TestGetHumanReadableVersionSemanticVersion tests semantic version output
func TestGetHumanReadableVersionSemanticVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "v1.0.0"
	result := GetHumanReadableVersion()
	expected := "gpsp-bot v1.0.0"
	if !strings.HasPrefix(result, expected) {
		t.Errorf("Expected result to start with %q, got %q", expected, result)
	}
}

// TestGetHumanReadableVersionContainsVersion ensures the actual version tag is visible
func TestGetHumanReadableVersionContainsVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	testCases := []string{"dev", "20231128151500", "v1.0.0", "v2.1.3-beta"}
	for _, version := range testCases {
		Version = version
		result := GetHumanReadableVersion()
		if !strings.Contains(result, version) {
			t.Errorf("Expected result to contain %q, got %q", version, result)
		}
	}
}

// TestGetHumanReadableVersionIncludesToolVersions verifies yt-dlp and ffmpeg info is present
func TestGetHumanReadableVersionIncludesToolVersions(t *testing.T) {
	result := GetHumanReadableVersion()
	if !strings.Contains(result, "yt-dlp:") {
		t.Errorf("Expected version output to contain yt-dlp version, got %q", result)
	}
	if !strings.Contains(result, "ffmpeg:") {
		t.Errorf("Expected version output to contain ffmpeg version, got %q", result)
	}
}
