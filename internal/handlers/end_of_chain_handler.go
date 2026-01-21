package handlers

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

const (
	cleanupMaxAge = 30 * 24 * time.Hour // 30 days
)

type EndOfChainHandler struct{}

func (h *EndOfChainHandler) Execute(m *Context) {
	slog.Debug("Entering EndOfChainHandler")
	if m.doneTyping != nil {
		slog.Debug("Closing doneTyping channel")
		close(m.doneTyping)
	}
	if m.action == DownloadVideo {
		utils.CleanupTmpDir(config.FromEnv().YTDLP_TMP_DIR)

		// Cleanup old fingerprints
		cfg := config.FromEnv()
		dbPath := filepath.Join(cfg.REPOST_DB_DIR, "repost_fingerprints.duckdb")
		if err := utils.CleanupOldFingerprints(dbPath, cleanupMaxAge); err != nil {
			slog.Warn("Failed to cleanup old fingerprints", "error", err)
		}
	}
	if m.action == Euribor {
		utils.CleanupTmpDir(config.FromEnv().EURIBOR_GRAPH_DIR)
		utils.CleanupTmpDir(config.FromEnv().EURIBOR_CSV_DIR)
	}

}

func (h *EndOfChainHandler) SetNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}
