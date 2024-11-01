package main

import "log"

type TelegramTelebotOnTextHandler struct {
	next ContextHandler
}

func (telegramMessageParser *TelegramTelebotOnTextHandler) execute(m *Context) {
	log.Println("Entering TelegramTelebotOnTextHandler")
	c := m.telebotContext
	message := c.Message()
	if message != nil {
		m.rawText = c.Message().Text
		m.id = c.Message().ID
		m.isReply = c.Message().IsReply()
		m.chatId = c.Chat().ID

		if (c.Message().IsReply()) {
			m.replyToId = c.Message().ReplyTo.ID
		}
	}
	telegramMessageParser.next.execute(m)
}

func (mp *TelegramTelebotOnTextHandler) setNext(next ContextHandler) {
	mp.next = next
}
