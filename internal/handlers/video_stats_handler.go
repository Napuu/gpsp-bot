package handlers

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

const statsDBFileName = "repost_fingerprints.duckdb"

type VideoStatsHandler struct {
	next ContextHandler
}

func (h *VideoStatsHandler) Execute(m *Context) {
	slog.Debug("Entering VideoStatsHandler")

	if m.sendVideoSucceeded && m.action == DownloadVideo && m.botMessageId != "" {
		cfg := config.FromEnv()
		dbPath := filepath.Join(cfg.REPOST_DB_DIR, statsDBFileName)

		var platform string
		switch m.Service {
		case Telegram:
			platform = "telegram"
		case Discord:
			platform = "discord"
		default:
			slog.Warn("VideoStatsHandler: unknown service, skipping", "service", m.Service)
			h.next.Execute(m)
			return
		}

		db, err := utils.OpenStatsDB(dbPath)
		if err != nil {
			slog.Warn("Failed to open stats DB", "error", err)
			h.next.Execute(m)
			return
		}
		defer db.Close()

		entry := utils.VideoStatEntry{
			Platform:     platform,
			GroupId:      platform + ":" + m.chatId,
			UserId:       m.posterUserId,
			Username:     m.posterUsername,
			SourceUrl:    m.url,
			BotMessageId: m.botMessageId,
			IsRepost:     m.isRepost,
			PostedAt:     time.Now(),
		}
		if err := utils.RecordVideoPost(db, entry); err != nil {
			slog.Warn("Failed to record video stat", "error", err)
		} else {
			slog.Debug("Video stat recorded", "platform", platform, "user", m.posterUsername, "url", m.url)
		}
	}

	h.next.Execute(m)
}

func (h *VideoStatsHandler) SetNext(next ContextHandler) {
	h.next = next
}
