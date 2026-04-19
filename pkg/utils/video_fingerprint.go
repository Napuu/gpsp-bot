package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

// ExtractOCRText extracts text from a video frame using tesseract.
// Returns the extracted text, whether the result is high-confidence, and any error.
// If tesseract is not installed, returns ("", false, nil) for graceful degradation.
func ExtractOCRText(videoPath, tmpDir string) (string, bool, error) {
	// Check if tesseract is available
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", false, nil
	}

	// Get video duration via ffprobe
	durationCmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	var durationOut bytes.Buffer
	durationCmd.Stdout = &durationOut
	durationCmd.Stderr = nil
	if err := durationCmd.Run(); err != nil {
		return "", false, fmt.Errorf("ffprobe duration failed: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(durationOut.String()), 64)
	if err != nil || duration <= 0 {
		return "", false, nil
	}

	midpoint := duration / 2.0

	// Extract one frame at midpoint (640px wide for OCR legibility)
	framePath := filepath.Join(tmpDir, "ocr_frame.png")
	extractCmd := exec.Command("ffmpeg",
		"-y",
		"-ss", fmt.Sprintf("%.3f", midpoint),
		"-i", videoPath,
		"-frames:v", "1",
		"-vf", "scale=640:-1",
		framePath,
	)
	extractCmd.Stderr = nil
	if err := extractCmd.Run(); err != nil {
		return "", false, fmt.Errorf("ffmpeg frame extract failed: %w", err)
	}

	// Run tesseract with TSV output for confidence data
	tsvCmd := exec.Command("tesseract", framePath, "stdout", "--psm", "6", "tsv")
	var tsvOut bytes.Buffer
	tsvCmd.Stdout = &tsvOut
	tsvCmd.Stderr = nil
	if err := tsvCmd.Run(); err != nil {
		return "", false, fmt.Errorf("tesseract failed: %w", err)
	}

	// Parse TSV output: filter words with conf >= 80
	var words []string
	allConfident := true
	for i, line := range strings.Split(tsvOut.String(), "\n") {
		if i == 0 { // skip header
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue
		}
		text := strings.TrimSpace(fields[11])
		if text == "" || text == "-1" {
			continue
		}
		conf, err := strconv.Atoi(strings.TrimSpace(fields[10]))
		if err != nil {
			continue
		}
		if conf < 80 {
			allConfident = false
			continue
		}
		words = append(words, text)
	}

	normalized := NormalizeOCRText(strings.Join(words, " "))

	// Require all words confident and total text >= 3 chars
	confident := allConfident && len(words) > 0 && len(normalized) >= 3
	if !confident {
		return "", false, nil
	}

	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:]), true, nil
}

// NormalizeOCRText lowercases, collapses whitespace, and trims OCR text.
func NormalizeOCRText(text string) string {
	text = strings.ToLower(text)
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}
