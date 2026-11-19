package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMainVersionFlag tests the -v flag by re-executing the test binary.
func TestMainVersionFlag(t *testing.T) {
	if os.Getenv("BE_MAIN") == "1" {
		os.Args = append([]string{"gpsp-bot"}, os.Args[len(os.Args)-1])
		main()
		return
	}

	testCases := []string{"-v", "--version"}

	for _, flag := range testCases {
		t.Run(flag, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestMainVersionFlag", "--", flag)
			cmd.Env = append(os.Environ(), "BE_MAIN=1")

			var combined bytes.Buffer
			cmd.Stdout = &combined
			cmd.Stderr = &combined

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Process ran with error: %v, output: %s", err, combined.String())
			}

			output := combined.String()
			if !strings.Contains(output, "gpsp-bot dev") {
				t.Errorf("Expected output to contain 'gpsp-bot dev', got %q", output)
			}
			if !strings.Contains(output, "yt-dlp:") {
				t.Errorf("Expected output to contain 'yt-dlp:', got %q", output)
			}
			if !strings.Contains(output, "ffmpeg:") {
				t.Errorf("Expected output to contain 'ffmpeg:', got %q", output)
			}
		})
	}
}

// TestMainNoArgs tests that running with no args exits with an error (usage).
func TestMainNoArgs(t *testing.T) {
	if os.Getenv("BE_MAIN") == "1" {
		os.Args = []string{"gpsp-bot"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainNoArgs")
	cmd.Env = append(os.Environ(), "BE_MAIN=1")
	// No additional args

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected process to exit with error, but it succeeded")
	}

	output := stderr.String()
	if !strings.Contains(output, "Usage: gpsp-bot") {
		t.Errorf("Expected output to contain 'Usage: gpsp-bot', got %q", output)
	}
}
