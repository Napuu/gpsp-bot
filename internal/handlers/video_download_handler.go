package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

type VideoDownloadHandler struct {
	next ContextHandler
	// Downloader is an optional override for the video download function.
	// It must return the local file path of the downloaded video, or an empty
	// string on failure.  When nil, utils.DownloadVideo is used.
	Downloader func(url string, targetSizeInMB uint64) string
}

func (u *VideoDownloadHandler) Execute(m *Context) {
	slog.Debug("Entering VideoDownloadHandler")
	if m.action == DownloadVideo {
		var videoString = m.url
		var path string
		if u.Downloader != nil {
			path = u.Downloader(videoString, 5)
		} else {
			path = utils.DownloadVideo(videoString, 5)
		}

		m.originalVideoPath = path
		m.finalVideoPath = path
	}
	u.next.Execute(m)
}

func (u *VideoDownloadHandler) SetNext(next ContextHandler) {
	u.next = next
}
