package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// GetRatesFromCSV must never crash the process when the csv directory does not
// exist yet. It should create the directory and return an empty result.
func TestGetRatesFromCSVCreatesMissingDir(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "euribor-exports")

	if _, err := os.Stat(missingDir); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist before the call", missingDir)
	}

	data := GetRatesFromCSV(missingDir, time.Now().AddDate(0, -1, 0))

	if len(data) != 0 {
		t.Fatalf("expected empty result for missing dir, got %d entries", len(data))
	}

	info, err := os.Stat(missingDir)
	if err != nil {
		t.Fatalf("expected directory to be created, stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", missingDir)
	}
}
