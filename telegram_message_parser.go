package main


type TelegramMessageParser struct {
	next handler
}

func (telegramMessageParser *TelegramMessageParser) execute(m *GenericMessage) {
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

func (mp *TelegramMessageParser) setNext(next handler) {
	mp.next = next
}
