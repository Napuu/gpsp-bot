package handlers

import (
	"log/slog"
	"strconv"
)

type TelegramTelebotOnTextHandler struct {
	next ContextHandler
}

func (telegramMessageParser *TelegramTelebotOnTextHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramTelebotOnTextHandler")
	if m.Service == Telegram {
		c := m.TelebotContext
		message := c.Message()
		if message != nil {
			m.rawText = c.Message().Text
			m.id = strconv.Itoa(c.Message().ID)
			m.isReply = c.Message().IsReply()
			m.chatId = strconv.Itoa(int(c.Chat().ID))

			if c.Message().IsReply() {
				m.replyToId = string(rune(c.Message().ReplyTo.ID))
				m.shouldReplyToMessage = true
			}
		}
	}
	telegramMessageParser.next.Execute(m)
}

func (mp *TelegramTelebotOnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
