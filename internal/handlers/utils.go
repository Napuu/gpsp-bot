package handlers

import (
	"context"
	"log/slog"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	tele "gopkg.in/telebot.v4"
)

func (m Context) SendTyping() {
	var err error

	switch m.Service {
	case Telegram:
		action := tele.Typing
		if m.action == DownloadVideo {
			action = tele.UploadingVideo
		}
		err = m.TelebotContext.Notify(action)
	case Discord:
		err = m.DiscordSession.ChannelTyping(m.chatId)
	case Matrix:
		if m.MatrixClient != nil && *m.MatrixClient != nil {
			client := (*m.MatrixClient).(*mautrix.Client)
			_, err = client.UserTyping(context.Background(), id.RoomID(m.chatId), true, 5*time.Second)
		}
	}

	if err != nil {
		slog.Debug("Failed to send typing indicator", "error", err)
	}
}
