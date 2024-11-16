package handlers

import (
	"log/slog"
	"strings"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

type VideoCutArgsHandler struct {
	next ContextHandler
}

func (u *VideoCutArgsHandler) Execute(m *Context) {
	slog.Debug("Entering VideoCutArgsHandler")
	leftover := strings.Replace(m.parsedText, m.url, "", 1)

	m.startSeconds = make(chan float64)
	m.durationSeconds = make(chan float64)

	MIN_LEFTOVER_LEN_TO_CONSIDER := 2
	m.cutVideoArgsParsed = make(chan bool)
	go func() {
		if m.action == DownloadVideo && len(leftover) > MIN_LEFTOVER_LEN_TO_CONSIDER {
			startSeconds, durationSeconds, err := utils.ParseCutArgs(leftover)
			if err != nil {
				slog.Error(err.Error())
				m.cutVideoArgsParsed <- false
			} else {
				m.cutVideoArgsParsed <- true
				m.startSeconds <- startSeconds
				m.durationSeconds <- durationSeconds
			}
		} else {
			m.cutVideoArgsParsed <- false
		}
	}()

	u.next.Execute(m)
}

func (u *VideoCutArgsHandler) SetNext(next ContextHandler) {
	u.next = next
}
