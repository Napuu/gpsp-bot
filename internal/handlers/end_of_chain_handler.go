package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type EndOfChainHandler struct{}

func (h *EndOfChainHandler) Execute(m *Context) {
	slog.Debug("Entering EndOfChainHandler")
	if m.doneTyping != nil {
		slog.Debug("Closing doneTyping channel")
		close(m.doneTyping)
	}

	utils.CleanupTmpDir(config.FromEnv().YTDLP_TMP_DIR)

}

func (h *EndOfChainHandler) SetNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}
