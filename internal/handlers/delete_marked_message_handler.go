package handlers

import (
	"context"
	"log/slog"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
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
		case Matrix:
			if m.MatrixClient != nil && *m.MatrixClient != nil {
				client := (*m.MatrixClient).(*mautrix.Client)
				_, err = client.RedactEvent(context.Background(), id.RoomID(m.chatId), id.EventID(m.id))
			}
		}

		if err != nil {
			slog.Warn("Failed to delete message", "error", err, "service", m.Service)
		}
	}

	r.next.Execute(m)
}

func (u *DeleteMessageHandler) SetNext(next ContextHandler) {
	u.next = next
}
