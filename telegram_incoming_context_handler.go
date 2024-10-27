package main


type TelegramIncomingContextHandler struct {
	next ContextHandler
}

func (telegramMessageParser *TelegramIncomingContextHandler) execute(m *Context) {
	c := m.telebotContext
	m.rawText = c.Message().Text
	m.id = c.Message().ID
	m.isReply = c.Message().IsReply()
	m.chatId = c.Chat().ID

	if (c.Message().IsReply()) {
		m.replyToId = c.Message().ReplyTo.ID
	}

	telegramMessageParser.next.execute(m)
}

func (mp *TelegramIncomingContextHandler) setNext(next ContextHandler) {
	mp.next = next
}
