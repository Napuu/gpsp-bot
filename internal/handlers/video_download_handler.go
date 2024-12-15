package handlers

import (
	"fmt"
	"log/slog"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

type VideoDownloadHandler struct {
	next ContextHandler
}

func (u *VideoDownloadHandler) Execute(m *Context) {
	slog.Debug("Entering VideoDownloadHandler")
	if m.action == DownloadVideo || m.action == SearchVideo {
		var videoString = m.url
		if m.action == SearchVideo {
			videoString = fmt.Sprintf("ytsearch:\"%s\"", m.parsedText)
		}
		path := utils.DownloadVideo(videoString, 5)

		m.originalVideoPath = path
		m.finalVideoPath = path
	}
	u.next.Execute(m)
}

func (u *VideoDownloadHandler) SetNext(next ContextHandler) {
	u.next = next
}
