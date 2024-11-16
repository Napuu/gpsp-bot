package handlers

import "log/slog"

type TelegramTelebotOnTextHandler struct {
	next ContextHandler
}

func (telegramMessageParser *TelegramTelebotOnTextHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramTelebotOnTextHandler")
	c := m.TelebotContext
	message := c.Message()
	if message != nil {
		m.rawText = c.Message().Text
		m.id = int64(c.Message().ID)
		m.isReply = c.Message().IsReply()
		m.chatId = c.Chat().ID

		if c.Message().IsReply() {
			m.replyToId = int64(c.Message().ReplyTo.ID)
			m.shouldReplyToMessage = true
		}
	}
	telegramMessageParser.next.Execute(m)
}

func (mp *TelegramTelebotOnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
