package handlers

import (
	"log"
	"log/slog"
)

type TelegramDeleteMarkedMessageHandler struct {
	next ContextHandler
}

func (r *TelegramDeleteMarkedMessageHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramDeleteMarkedMessageHandler")
	if m.Service == Telegram && m.shouldDeleteOriginalMessage {
		err := m.TelebotContext.Delete()
		if err != nil {
			log.Println(err)
		}
	}

	r.next.Execute(m)
}

func (u *TelegramDeleteMarkedMessageHandler) SetNext(next ContextHandler) {
	u.next = next
}
