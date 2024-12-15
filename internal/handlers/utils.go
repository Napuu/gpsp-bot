package handlers

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

func (m Context) SendTyping() {
	var err error

	switch m.Service {
	case Telegram:
		action := tele.Typing
		if m.action == DownloadVideo || m.action == SearchVideo {
			action = tele.UploadingVideo
		}
		err = m.TelebotContext.Notify(action)
	case Discord:
		err = m.DiscordSession.ChannelTyping(m.chatId)
	}

	if err != nil {
		slog.Error(err.Error())
	}
}
