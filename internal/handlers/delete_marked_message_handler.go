package handlers

import (
	"log/slog"
)

type DeleteMessageHandler struct {
	next ContextHandler
}

func (r *DeleteMessageHandler) Execute(m *Context) {
	slog.Debug("Entering DeleteMarkedMessageHandler")

	if m.shouldDeleteOriginalMessage {
		var err error

		switch m.Service {
		case Telegram:
			err = m.TelebotContext.Delete()
		case Discord:
			err = m.DiscordSession.ChannelMessageDelete(m.chatId, m.id)
		}

		if err != nil {
			slog.Warn(err.Error())
		}
	}

	r.next.Execute(m)
}

func (u *DeleteMessageHandler) SetNext(next ContextHandler) {
	u.next = next
}
