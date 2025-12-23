package utils

import (
	"testing"
)

func TestIsYleURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "yle.fi direct domain",
			url:      "https://yle.fi/aihe/artikkeli/2023/12/01/video",
			expected: true,
		},
		{
			name:     "areena.yle.fi subdomain",
			url:      "https://areena.yle.fi/1-12345678",
			expected: true,
		},
		{
			name:     "yle.fi with www",
			url:      "https://www.yle.fi/video/12345",
			expected: true,
		},
		{
			name:     "youtube url",
			url:      "https://www.youtube.com/watch?v=12345",
			expected: false,
		},
		{
			name:     "generic url",
			url:      "https://example.com/video",
			expected: false,
		},
		{
			name:     "yle.fi in path but not domain",
			url:      "https://example.com/yle.fi/video",
			expected: true, // This will match, but that's okay - false positives are safe since we fallback to yt-dlp
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isYleURL(tt.url)
			if result != tt.expected {
				t.Errorf("isYleURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}
