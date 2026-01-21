package utils

import (
	"bytes"
	"math"
	"os/exec"
)

// GetVideoFingerprint returns a slice of bytes representing the video structure
// using low-resolution temporal grid (8x8 grayscale, 1fps)
func GetVideoFingerprint(filePath string) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-vf", "fps=1,scale=8:8:flags=lanczos,format=gray",
		"-map", "0:v:0",
		"-f", "rawvideo",
		"-",
	)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = nil // Suppress ffmpeg logs

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

// CalculateSimilarity returns a score between 0.0 (different) and 1.0 (identical)
// The score is penalized based on length differences to avoid false positives
// when short clips share opening frames with longer videos.
func CalculateSimilarity(fp1, fp2 []byte) float64 {
	// Simple length check
	if len(fp1) == 0 || len(fp2) == 0 {
		return 0.0
	}

	// Truncate to the shorter length to compare overlaps
	minLen := len(fp1)
	maxLen := len(fp1)
	if len(fp2) < minLen {
		minLen = len(fp2)
	}
	if len(fp2) > maxLen {
		maxLen = len(fp2)
	}

	diffSum := 0.0

	// Compare byte by byte (pixel by pixel)
	for i := 0; i < minLen; i++ {
		// Calculate absolute difference between pixel values (0-255)
		val1 := float64(fp1[i])
		val2 := float64(fp2[i])
		diffSum += math.Abs(val1 - val2)
	}

	// Average difference per pixel
	avgDiff := diffSum / float64(minLen)

	// Normalize to 0..1 score.
	// Max difference per pixel is 255.
	// If average difference is 0, score is 1.0.
	// If average difference is > 64, score drops.
	overlapScore := 1.0 - (avgDiff / 64.0)
	if overlapScore < 0 {
		overlapScore = 0
	}

	// Apply length ratio penalty to avoid false positives when videos
	// have significantly different durations. A 10-second clip matching
	// the first 10 seconds of a 60-second video should not be considered
	// a repost (ratio = 10/60 = 0.167).
	lengthRatio := float64(minLen) / float64(maxLen)
	score := overlapScore * lengthRatio

	return score
}
