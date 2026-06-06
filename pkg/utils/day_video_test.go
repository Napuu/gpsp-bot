package utils

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOrdinalSuffix(t *testing.T) {
	tests := []struct {
		day      int
		expected string
	}{
		{1, "st"},
		{2, "nd"},
		{3, "rd"},
		{4, "th"},
		{11, "th"},
		{12, "th"},
		{13, "th"},
		{21, "st"},
		{22, "nd"},
		{23, "rd"},
		{31, "st"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := ordinalSuffix(tt.day); got != tt.expected {
				t.Errorf("ordinalSuffix(%d) = %q, want %q", tt.day, got, tt.expected)
			}
		})
	}
}

func TestFormatDayOverlayText(t *testing.T) {
	tm := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)
	dateLine, weekdayLine := FormatDayOverlayText(tm)

	if dateLine != "6th June, 2026" {
		t.Errorf("dateLine = %q, want %q", dateLine, "6th June, 2026")
	}
	if weekdayLine != "Saturday" {
		t.Errorf("weekdayLine = %q, want %q", weekdayLine, "Saturday")
	}
}

func TestFormatDayOverlayTextLongestMonth(t *testing.T) {
	tm := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	dateLine, weekdayLine := FormatDayOverlayText(tm)

	if dateLine != "30th September, 2026" {
		t.Errorf("dateLine = %q, want %q", dateLine, "30th September, 2026")
	}
	if weekdayLine != "Wednesday" {
		t.Errorf("weekdayLine = %q, want %q", weekdayLine, "Wednesday")
	}
}

func TestGenerateDayTextPNG(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "day-text-*.png")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := GenerateDayTextPNG("6th June, 2026", "Saturday", tmpPath); err != nil {
		t.Fatalf("GenerateDayTextPNG() error: %v", err)
	}

	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("stat png: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("png file is empty")
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		t.Fatalf("open png: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	if !imageHasOpaquePixels(img) {
		t.Error("png has no opaque pixels")
	}
}

func imageHasOpaquePixels(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				return true
			}
		}
	}
	return false
}

func TestGenerateDayVideo(t *testing.T) {
	if !isCommandAvailable("ffmpeg") {
		t.Skip("ffmpeg not available, skipping test")
	}

	outputPath := filepath.Join(t.TempDir(), "day-video.mp4")
	tm := time.Date(2026, time.June, 6, 12, 0, 0, 0, time.UTC)

	if err := GenerateDayVideo(outputPath, tm); err != nil {
		t.Fatalf("GenerateDayVideo() error: %v", err)
	}

	if !isValidVideoFile(outputPath) {
		t.Fatalf("output is not a valid video file: %s", outputPath)
	}
}
