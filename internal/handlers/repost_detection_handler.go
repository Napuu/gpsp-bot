package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

const (
	similarityThreshold = 0.95
)

type RepostDetectionHandler struct {
	next ContextHandler
}

func (r *RepostDetectionHandler) Execute(m *Context) {
	slog.Debug("Entering RepostDetectionHandler")

	if m.action != DownloadVideo || m.finalVideoPath == "" {
		r.next.Execute(m)
		return
	}

	cfg := config.FromEnv()
	dbDir := cfg.REPOST_DB_DIR
	dbPath := filepath.Join(dbDir, "repost_fingerprints.duckdb")

	// Initialize database if needed
	if err := utils.InitRepostDB(dbPath); err != nil {
		slog.Warn("Failed to initialize repost database", "error", err)
		r.next.Execute(m)
		return
	}

	// Generate fingerprint
	fingerprint, err := utils.GetVideoFingerprint(m.finalVideoPath)
	if err != nil {
		slog.Warn("Failed to generate video fingerprint", "error", err)
		// Continue without fingerprinting - don't block video
		r.next.Execute(m)
		return
	}

	// Extract OCR text hash from a middle frame
	tmpDir := config.FromEnv().YTDLP_TMP_DIR
	ocrHash, ocrConfident, ocrErr := utils.ExtractOCRText(m.finalVideoPath, tmpDir)
	if ocrErr != nil {
		slog.Warn("OCR extraction failed", "error", ocrErr)
	}
	if ocrConfident {
		m.pendingOCRText = ocrHash
		slog.Debug("OCR text hash extracted", "hash", ocrHash)
	}

	// Build platform-prefixed group_id
	var platformPrefix string
	switch m.Service {
	case Telegram:
		platformPrefix = "telegram"
	case Discord:
		platformPrefix = "discord"
	default:
		slog.Warn("Unknown service type, skipping repost detection")
		r.next.Execute(m)
		return
	}

	groupId := fmt.Sprintf("%s:%s", platformPrefix, m.chatId)

	// Query for similar fingerprints within the same group
	matches, err := utils.FindSimilarFingerprints(dbPath, fingerprint, groupId, similarityThreshold)
	if err != nil {
		slog.Warn("Failed to query for similar fingerprints", "error", err)
		// Continue without detection - don't block video
		// Store fingerprint data in context to be stored after video message is sent
		// We need the message ID of the bot's response, not the original message
		m.pendingFingerprint = fingerprint
		m.pendingFingerprintGroupId = groupId
		m.pendingFingerprintDbPath = dbPath
		slog.Debug("Fingerprint data prepared for storage after message is sent", "groupId", groupId)
	} else if len(matches) > 0 {
		// Filter matches using OCR: if both videos have OCR text but text differs, skip match
		var filteredMessageIds []string
		for _, match := range matches {
			if m.pendingOCRText != "" && match.OCRTextHash != "" && m.pendingOCRText != match.OCRTextHash {
				slog.Info("OCR text hash differs, skipping visual match", "newHash", m.pendingOCRText, "existingHash", match.OCRTextHash, "messageId", match.MessageID)
				continue
			}
			filteredMessageIds = append(filteredMessageIds, match.MessageID)
		}

		if len(filteredMessageIds) == 0 {
			// All matches were filtered out by OCR — not a repost
			m.pendingFingerprint = fingerprint
			m.pendingFingerprintGroupId = groupId
			m.pendingFingerprintDbPath = dbPath
			slog.Debug("All visual matches overridden by OCR text difference", "groupId", groupId)
			r.next.Execute(m)
			return
		}

		// Repost detected
		m.isRepost = true
		m.repostOriginalMessageIds = filteredMessageIds
		slog.Info("Repost detected", "groupId", groupId, "matches", len(filteredMessageIds))

		// Generate composite image with first frame
		if m.finalVideoPath != "" {
			if err := r.generateRepostImage(m); err != nil {
				slog.Warn("Failed to generate repost image", "error", err)
				// Continue without image - fall back to text only
			}
		}
		// Do not store fingerprint for reposts to avoid duplicate entries
	} else {
		// No repost detected - store fingerprint data in context to be stored after video message is sent
		// We need the message ID of the bot's response, not the original message
		m.pendingFingerprint = fingerprint
		m.pendingFingerprintGroupId = groupId
		m.pendingFingerprintDbPath = dbPath
		slog.Debug("Fingerprint data prepared for storage after message is sent", "groupId", groupId)
	}

	r.next.Execute(m)
}

func (r *RepostDetectionHandler) SetNext(next ContextHandler) {
	r.next = next
}

func (r *RepostDetectionHandler) generateRepostImage(m *Context) error {
	// Generate temporary paths
	frameID := uuid.New().String()
	compositeID := uuid.New().String()
	tmpDir := config.FromEnv().YTDLP_TMP_DIR
	framePath := fmt.Sprintf("%s/%s_frame.jpg", tmpDir, frameID)
	compositePath := fmt.Sprintf("%s/%s_composite.png", tmpDir, compositeID)

	// Extract first frame from video
	if err := utils.ExtractFirstFrame(m.finalVideoPath, framePath); err != nil {
		return fmt.Errorf("failed to extract first frame: %w", err)
	}

	// Get template image
	templateImg, err := utils.GetTemplateImage()
	if err != nil {
		return fmt.Errorf("failed to get template image: %w", err)
	}

	// Composite the frame into the template
	if err := utils.CompositeRepostImage(templateImg, framePath, compositePath); err != nil {
		// Clean up frame file even if composite fails
		os.Remove(framePath)
		return fmt.Errorf("failed to composite image: %w", err)
	}

	// Clean up the intermediate frame file - it's no longer needed after composite is created
	if err := os.Remove(framePath); err != nil {
		slog.Warn("Failed to delete intermediate frame file", "path", framePath, "error", err)
		// Continue anyway - this is not critical
	}

	// Set the composite image path and text response
	m.finalImagePath = compositePath
	m.textResponse = "️️️❗❗❗⚠️⚠️⚠️❗❗❗\nReposti tunnistettu!\n️❗❗❗⚠️⚠️⚠️❗❗❗"
	// Use image-specific reply fields so video response handler can use original reply target
	m.imageShouldReplyToMessage = true
	m.imageReplyToId = m.repostOriginalMessageIds[0]

	return nil
}
