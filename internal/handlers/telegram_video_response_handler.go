package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type TelegramVideoResponseHandler struct {
	next ContextHandler
}

func (r *TelegramVideoResponseHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramVideoResponseHandler")
	if m.Service == Telegram {
		chatId := tele.ChatID(utils.S2I(m.chatId))

		var videoPathToSend = m.originalVideoPath
		if len(m.possiblyProcessedVideoPath) > 0 {
			videoPathToSend = m.possiblyProcessedVideoPath
		}

		if len(videoPathToSend) > 0 {
			m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(videoPathToSend)})
			m.sendVideoSucceeded = true
		}
	}

	r.next.Execute(m)
}

func (u *TelegramVideoResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
