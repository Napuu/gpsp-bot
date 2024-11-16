package handlers

import "log/slog"

type MarkForDeletionHandler struct {
	next ContextHandler
}

func (u *MarkForDeletionHandler) Execute(m *Context) {
	slog.Debug("Entering MarkForDeletionHandler")
	if (m.action == DownloadVideo || m.action == SearchVideo) && m.sendVideoSucceeded {
		m.shouldDeleteOriginalMessage = true
	}

	u.next.Execute(m)
}

func (u *MarkForDeletionHandler) SetNext(next ContextHandler) {
	u.next = next
}
