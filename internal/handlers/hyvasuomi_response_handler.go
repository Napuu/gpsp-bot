package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type HyvaSuomiResponseHandler struct {
	next ContextHandler
}

func (r *HyvaSuomiResponseHandler) Execute(m *Context) {
	slog.Debug("Entering HyvaSuomiResponseHandler")

	if m.action == Tuplilla && m.Service == Telegram && m.gotHyvaSuomi {
		chatId := tele.ChatID(utils.S2I(m.chatId))
		m.Telebot.Send(chatId, "Hyvä suomi 🇫🇮🇫🇮")
	}

	r.next.Execute(m)
}

func (u *HyvaSuomiResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}