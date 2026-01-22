package utils

import (
	"image"
	"image/color"
	"testing"

	"github.com/disintegration/imaging"
)

func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0,
		},
		{
			name:     "all same values",
			values:   []float64{5, 5, 5, 5, 5},
			expected: 0,
		},
		{
			name:     "different values",
			values:   []float64{1, 2, 3, 4, 5},
			expected: 2.0, // variance of [1,2,3,4,5] is 2.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVariance(tt.values)
			if result != tt.expected {
				t.Errorf("calculateVariance(%v) = %v, expected %v", tt.values, result, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{
			name:     "zero seconds",
			seconds:  0.0,
			expected: "0.000",
		},
		{
			name:     "half second",
			seconds:  0.5,
			expected: "0.500",
		},
		{
			name:     "one second",
			seconds:  1.0,
			expected: "1.000",
		},
		{
			name:     "fractional seconds",
			seconds:  1.234,
			expected: "1.234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatTime(%v) = %v, expected %v", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestIsFrameSingleColor(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Create a temporary black image
	blackImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			blackImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	// Create a temporary colorful image
	colorfulImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Create a gradient effect
			r := uint8(x * 255 / 100)
			g := uint8(y * 255 / 100)
			b := uint8((x + y) * 255 / 200)
			colorfulImg.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Save test images to temp files
	blackPath := tmpDir + "/test_black_frame.jpg"
	colorfulPath := tmpDir + "/test_colorful_frame.jpg"

	if err := imaging.Save(blackImg, blackPath); err != nil {
		t.Fatalf("Failed to save black test image: %v", err)
	}

	if err := imaging.Save(colorfulImg, colorfulPath); err != nil {
		t.Fatalf("Failed to save colorful test image: %v", err)
	}

	tests := []struct {
		name      string
		imagePath string
		expected  bool
	}{
		{
			name:      "black single-color frame",
			imagePath: blackPath,
			expected:  true,
		},
		{
			name:      "colorful multi-color frame",
			imagePath: colorfulPath,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isFrameSingleColor(tt.imagePath)
			if err != nil {
				t.Fatalf("isFrameSingleColor(%q) returned error: %v", tt.imagePath, err)
			}
			if result != tt.expected {
				t.Errorf("isFrameSingleColor(%q) = %v, expected %v", tt.imagePath, result, tt.expected)
			}
		})
	}
}
