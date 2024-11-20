package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type TelegramTextResponseHandler struct {
	next ContextHandler
}

func (r *TelegramTextResponseHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramTextResponseHandler")
	if m.Service == Telegram {
		if m.shouldReplyToMessage {
			message := &tele.Message{
				Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
				ID:   utils.S2I(m.replyToId),
			}
			m.Telebot.Reply(message, m.textResponse)
		} else if m.textResponse != "" {
			m.TelebotContext.Send(m.textResponse)
		}
	}

	r.next.Execute(m)
}

func (u *TelegramTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
