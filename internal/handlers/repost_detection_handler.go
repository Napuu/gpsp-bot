package handlers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type RepostDetectionHandler struct {
	next            ContextHandler
	cleanupCounter  int
	cleanupMutex    sync.Mutex
	lastCleanupTime time.Time
}

const (
	similarityThreshold = 0.95
	cleanupInterval     = 10                  // Run cleanup every 10 detections
	cleanupMaxAge       = 30 * 24 * time.Hour // 30 days
)

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
	} else if len(matches) > 0 {
		// Repost detected
		m.isRepost = true
		m.repostOriginalMessageIds = matches
		slog.Info("Repost detected", "groupId", groupId, "matches", len(matches))

		// Generate composite image with first frame
		if m.finalVideoPath != "" {
			if err := r.generateRepostImage(m); err != nil {
				slog.Warn("Failed to generate repost image", "error", err)
				// Continue without image - fall back to text only
			}
		}
	}

	// Store fingerprint data in context to be stored after video message is sent
	// We need the message ID of the bot's response, not the original message
	m.pendingFingerprint = fingerprint
	m.pendingFingerprintGroupId = groupId
	m.pendingFingerprintDbPath = dbPath
	slog.Debug("Fingerprint data prepared for storage after message is sent", "groupId", groupId)

	// Periodic cleanup
	r.cleanupMutex.Lock()
	r.cleanupCounter++
	shouldCleanup := r.cleanupCounter >= cleanupInterval || time.Since(r.lastCleanupTime) > 24*time.Hour
	if shouldCleanup {
		r.cleanupCounter = 0
		r.lastCleanupTime = time.Now()
		r.cleanupMutex.Unlock()

		if err := utils.CleanupOldFingerprints(dbPath, cleanupMaxAge); err != nil {
			slog.Warn("Failed to cleanup old fingerprints", "error", err)
		}
	} else {
		r.cleanupMutex.Unlock()
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
		return fmt.Errorf("failed to composite image: %w", err)
	}

	// Set the composite image path and text response
	m.finalImagePath = compositePath
	m.textResponse = "️️️❗❗❗⚠️⚠️⚠️❗❗❗\nReposti tunnistettu!\n️❗❗❗⚠️⚠️⚠️❗❗❗"
	m.shouldReplyToMessage = true
	m.replyToId = m.repostOriginalMessageIds[0]

	return nil
}
