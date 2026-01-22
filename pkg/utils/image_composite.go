package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os/exec"

	"github.com/disintegration/imaging"
	"github.com/napuu/gpsp-bot/assets"
)

const (
	rectX      = 75
	rectY      = 135
	rectWidth  = 95
	rectHeight = 90

	// Threshold for single-color detection - variance below this is considered single-color
	colorVarianceThreshold = 10.0
	// Time to skip ahead if first frame is single-color (in seconds)
	fadeInSkipTime = 0.5
)

// ExtractFirstFrame extracts the first frame from a video using ffmpeg
// This function now intelligently avoids single-color frames (like black screens)
// by checking if the first frame is single-color and if so, extracting from 0.5s instead
func ExtractFirstFrame(videoPath, outputPath string) error {
	// First, extract frame at time 0
	if err := extractFrameAtTime(videoPath, outputPath, 0.0); err != nil {
		return err
	}

	// Check if the extracted frame is single-color
	isSingleColor, err := isFrameSingleColor(outputPath)
	if err != nil {
		// If we can't check, just use the frame we have
		return nil
	}

	// If the first frame is single-color (e.g., black screen during fade-in),
	// extract a frame from 0.5 seconds instead to skip the fade
	if isSingleColor {
		if err := extractFrameAtTime(videoPath, outputPath, fadeInSkipTime); err != nil {
			// If extraction at 0.5s fails, keep the original frame
			return nil
		}
	}

	return nil
}

// extractFrameAtTime extracts a frame from a video at a specific timestamp using ffmpeg
func extractFrameAtTime(videoPath, outputPath string, timeSeconds float64) error {
	args := []string{
		"-i", videoPath,
	}

	// Add seek time if not at the beginning
	if timeSeconds > 0 {
		args = append(args, "-ss", formatTime(timeSeconds))
	}

	args = append(args,
		"-vframes", "1",
		"-q:v", "2",
		outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// formatTime converts seconds to a time format for ffmpeg
func formatTime(seconds float64) string {
	return fmt.Sprintf("%.3f", seconds)
}

// isFrameSingleColor checks if an image frame is approximately a single color
// by sampling multiple points and calculating color variance
func isFrameSingleColor(imagePath string) (bool, error) {
	// Open and decode the image
	img, err := imaging.Open(imagePath)
	if err != nil {
		return false, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Sample points in a grid pattern (3x5 grid = 15 sample points)
	samplePoints := []struct{ x, y int }{
		// Top row
		{width / 4, height / 4},
		{width / 2, height / 4},
		{3 * width / 4, height / 4},
		// Middle-top row
		{width / 4, height / 3},
		{width / 2, height / 3},
		{3 * width / 4, height / 3},
		// Center row
		{width / 4, height / 2},
		{width / 2, height / 2},
		{3 * width / 4, height / 2},
		// Middle-bottom row
		{width / 4, 2 * height / 3},
		{width / 2, 2 * height / 3},
		{3 * width / 4, 2 * height / 3},
		// Bottom row
		{width / 4, 3 * height / 4},
		{width / 2, 3 * height / 4},
		{3 * width / 4, 3 * height / 4},
	}

	// Collect RGB values from sample points
	var rValues, gValues, bValues []float64
	for _, point := range samplePoints {
		c := img.At(point.x, point.y)
		r, g, b, _ := c.RGBA()
		// Convert from 16-bit to 8-bit color values
		rValues = append(rValues, float64(r>>8))
		gValues = append(gValues, float64(g>>8))
		bValues = append(bValues, float64(b>>8))
	}

	// Calculate variance for each channel
	rVariance := calculateVariance(rValues)
	gVariance := calculateVariance(gValues)
	bVariance := calculateVariance(bValues)

	// If all channels have low variance, the image is single-color
	maxVariance := math.Max(math.Max(rVariance, gVariance), bVariance)
	return maxVariance < colorVarianceThreshold, nil
}

// calculateVariance calculates the variance of a slice of float64 values
func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	return variance
}

// GetTemplateImage reads and decodes the embedded template image
func GetTemplateImage() (image.Image, error) {
	data, err := assets.TemplateFS.ReadFile("pate_template.jpg")
	if err != nil {
		return nil, err
	}

	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// CompositeRepostImage composites the frame into the template at the specified coordinates
func CompositeRepostImage(templateImg image.Image, framePath, outputPath string) error {
	// Load the frame image
	frameImg, err := imaging.Open(framePath)
	if err != nil {
		return err
	}

	// Resize frame to fit the rectangle
	resizedFrame := imaging.Resize(frameImg, rectWidth, rectHeight, imaging.Lanczos)

	// Create a mutable copy of the template
	result := imaging.Clone(templateImg)

	// Draw the resized frame onto the result
	targetRect := image.Rect(rectX, rectY, rectX+rectWidth, rectY+rectHeight)
	draw.Draw(result, targetRect, resizedFrame, image.Point{0, 0}, draw.Over)

	// Save the final image
	err = imaging.Save(result, outputPath)
	if err != nil {
		return err
	}

	return nil
}
