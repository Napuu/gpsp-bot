package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

// BotForwardIngestHandler seeds repost fingerprints when a user forwards
// this bot's video/animation into a Telegram group.
type BotForwardIngestHandler struct {
	next ContextHandler
}

func (h *BotForwardIngestHandler) Execute(m *Context) {
	slog.Debug("Entering BotForwardIngestHandler")

	if !shouldIngestBotForward(m) {
		h.next.Execute(m)
		return
	}

	msg := m.TelebotContext.Message()
	file := mediaFile(msg)
	if file == nil {
		h.next.Execute(m)
		return
	}

	cfg := config.FromEnv()
	tmpPath := filepath.Join(cfg.YTDLP_TMP_DIR, fmt.Sprintf("bot-forward-%s.mp4", uuid.New().String()))
	if err := os.MkdirAll(cfg.YTDLP_TMP_DIR, 0o755); err != nil {
		slog.Warn("Failed to create temp dir for bot forward ingest", "error", err)
		h.next.Execute(m)
		return
	}

	if err := m.Telebot.Download(file, tmpPath); err != nil {
		slog.Warn("Failed to download forwarded bot media", "error", err)
		h.next.Execute(m)
		return
	}
	defer func() {
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("Failed to remove temp forward media", "path", tmpPath, "error", err)
		}
	}()

	fingerprint, err := utils.GetVideoFingerprint(tmpPath)
	if err != nil {
		slog.Warn("Failed to fingerprint forwarded bot media", "error", err)
		h.next.Execute(m)
		return
	}

	ocrHash, ocrConfident, ocrErr := utils.ExtractOCRText(tmpPath, cfg.YTDLP_TMP_DIR)
	if ocrErr != nil {
		slog.Warn("OCR extraction failed for forwarded bot media", "error", ocrErr)
	}
	var pendingOCR string
	if ocrConfident {
		pendingOCR = ocrHash
	}

	dbPath := filepath.Join(cfg.REPOST_DB_DIR, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		slog.Warn("Failed to initialize repost database for forward ingest", "error", err)
		h.next.Execute(m)
		return
	}

	groupId := fmt.Sprintf("telegram:%s", m.chatId)
	matches, err := utils.FindSimilarFingerprints(dbPath, fingerprint, groupId, similarityThreshold)
	if err != nil {
		slog.Warn("Failed to query fingerprints for forward ingest", "error", err)
		h.next.Execute(m)
		return
	}

	if hasSurvivingMatch(matches, pendingOCR) {
		slog.Debug("Forwarded bot media already fingerprinted in group, skipping store", "groupId", groupId)
		h.next.Execute(m)
		return
	}

	messageId := m.id
	if err := utils.StoreFingerprint(dbPath, fingerprint, groupId, messageId, pendingOCR); err != nil {
		slog.Warn("Failed to store fingerprint for forwarded bot media", "error", err)
	} else {
		slog.Info("Stored fingerprint for forwarded bot media", "groupId", groupId, "messageId", messageId)
	}

	h.next.Execute(m)
}

func (h *BotForwardIngestHandler) SetNext(next ContextHandler) {
	h.next = next
}

func shouldIngestBotForward(m *Context) bool {
	if m.Service != Telegram || m.Telebot == nil || m.TelebotContext == nil {
		return false
	}
	if !slices.Contains(config.EnabledFeatures(), string(DownloadVideo)) {
		return false
	}
	msg := m.TelebotContext.Message()
	if msg == nil || !msg.FromGroup() {
		return false
	}
	if m.Telebot.Me == nil || !isForwardedFromBot(msg, m.Telebot.Me.ID) {
		return false
	}
	return mediaFile(msg) != nil
}

// isForwardedFromBot reports whether msg was forwarded from the given bot user.
// Prefer Origin (Bot API 7+) and fall back to legacy OriginalSender.
// IsForwarded() alone is insufficient — it ignores Origin.
func isForwardedFromBot(msg *tele.Message, botID int64) bool {
	if msg == nil {
		return false
	}
	if msg.Origin != nil && msg.Origin.Sender != nil && msg.Origin.Sender.ID == botID {
		return true
	}
	if msg.OriginalSender != nil && msg.OriginalSender.ID == botID {
		return true
	}
	return false
}

func mediaFile(msg *tele.Message) *tele.File {
	if msg == nil {
		return nil
	}
	if msg.Video != nil {
		return &msg.Video.File
	}
	if msg.Animation != nil {
		return &msg.Animation.File
	}
	return nil
}

// hasSurvivingMatch mirrors repost detection OCR filtering: a visual match is
// discarded when both sides have OCR hashes that differ.
func hasSurvivingMatch(matches []utils.FingerprintMatch, pendingOCR string) bool {
	for _, match := range matches {
		if pendingOCR != "" && match.OCRTextHash != "" && pendingOCR != match.OCRTextHash {
			continue
		}
		return true
	}
	return false
}
