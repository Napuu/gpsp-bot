package handlers

import (
	"log/slog"
)

type MarkForNaggingHandler struct {
	next ContextHandler
}

func (u *MarkForNaggingHandler) Execute(m *Context) {
	slog.Debug("Entering MarkForNaggingHandler")
	if m.action == DownloadVideo && !m.sendVideoSucceeded {
		slog.Debug("shouldNag set true")
		m.shouldNagAboutOriginalMessage = true
	}

	u.next.Execute(m)
}

func (u *MarkForNaggingHandler) SetNext(next ContextHandler) {
	u.next = next
}
