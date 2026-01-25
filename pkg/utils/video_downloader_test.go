package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsYleFiURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "root yle.fi",
			url:      "https://yle.fi/areena",
			expected: true,
		},
		{
			name:     "subdomain yle.fi",
			url:      "https://areena.yle.fi/video/123",
			expected: true,
		},
		{
			name:     "another subdomain",
			url:      "https://www.yle.fi/news",
			expected: true,
		},
		{
			name:     "non-yle.fi domain",
			url:      "https://youtube.com/watch?v=123",
			expected: false,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: false,
		},
		{
			name:     "yle.fi in path but not domain",
			url:      "https://example.com/yle.fi",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isYleFiURL(tt.url)
			if result != tt.expected {
				t.Errorf("isYleFiURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestGetSpecialExtractor(t *testing.T) {
	yleDlAvailable := isCommandAvailable("yle-dl")

	tests := []struct {
		name          string
		url           string
		expectedCmd   string
		expectedEmpty bool
		requiresYleDl bool
	}{
		{
			name:          "yle.fi URL with yle-dl available",
			url:           "https://areena.yle.fi/video/123",
			expectedCmd:   "yle-dl",
			expectedEmpty: !yleDlAvailable,
			requiresYleDl: true,
		},
		{
			name:          "root yle.fi URL",
			url:           "https://yle.fi/areena",
			expectedCmd:   "yle-dl",
			expectedEmpty: !yleDlAvailable,
			requiresYleDl: true,
		},
		{
			name:          "non-yle.fi URL",
			url:           "https://youtube.com/watch?v=123",
			expectedEmpty: true,
		},
		{
			name:          "invalid URL",
			url:           "not-a-url",
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := getSpecialExtractor(tt.url)

			if tt.expectedEmpty {
				if extractor.Command != "" {
					t.Errorf("getSpecialExtractor(%q) should return empty extractor, got Command=%q", tt.url, extractor.Command)
				}
			} else {
				if extractor.Command != tt.expectedCmd {
					t.Errorf("getSpecialExtractor(%q) Command = %q, expected %q", tt.url, extractor.Command, tt.expectedCmd)
				}
				if extractor.URLMatcher == nil {
					t.Error("URLMatcher should not be nil")
				}
				if extractor.DownloadFunc == nil {
					t.Error("DownloadFunc should not be nil")
				}
				if tt.requiresYleDl && extractor.Command == "yle-dl" {
					if extractor.SupportsProxy {
						t.Error("yle-dl extractor should not support proxy")
					}
				}
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected bool
	}{
		{
			name:     "existing command",
			cmd:      "go",
			expected: true,
		},
		{
			name:     "non-existent command",
			cmd:      "definitely-does-not-exist-command-xyz123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommandAvailable(tt.cmd)
			if result != tt.expected {
				t.Errorf("isCommandAvailable(%q) = %v, expected %v", tt.cmd, result, tt.expected)
			}
		})
	}
}

func TestAttemptHTTPDownload(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "invalid URL",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "non-existent domain",
			url:         "http://this-domain-definitely-does-not-exist-xyz123.com/video.mp4",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := attemptHTTPDownload(tt.url, "/tmp/test-video.mp4", "", 0)
			if !tt.expectError && !result {
				t.Errorf("attemptHTTPDownload(%q) failed but was expected to succeed", tt.url)
			}
			if tt.expectError && result {
				t.Errorf("attemptHTTPDownload(%q) succeeded but was expected to fail", tt.url)
			}
		})
	}
}

func TestIsValidVideoFile(t *testing.T) {
	// Check if ffprobe is available
	if !isCommandAvailable("ffprobe") {
		t.Skip("ffprobe not available, skipping test")
	}

	tests := []struct {
		name        string
		content     string
		expected    bool
		description string
	}{
		{
			name:        "non-existent file",
			content:     "",
			expected:    false,
			description: "File that doesn't exist should fail validation",
		},
		{
			name:        "HTML file",
			content:     "<html><body>Not a video</body></html>",
			expected:    false,
			description: "HTML file should fail validation",
		},
		{
			name:        "text file",
			content:     "This is just plain text, not a video file",
			expected:    false,
			description: "Text file should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.content != "" {
				// Create temporary file with the content
				tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				filePath = tmpFile.Name()
				defer os.Remove(filePath)

				if _, err := tmpFile.WriteString(tt.content); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				tmpFile.Close()
			} else {
				// Non-existent file
				filePath = filepath.Join(os.TempDir(), "definitely-does-not-exist-xyz123.mp4")
			}

			result := isValidVideoFile(filePath)
			if result != tt.expected {
				t.Errorf("%s: isValidVideoFile() = %v, expected %v", tt.description, result, tt.expected)
			}
		})
	}
}
