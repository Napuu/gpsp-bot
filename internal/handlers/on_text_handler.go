package handlers

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"maunium.net/go/mautrix/event"
)

type OnTextHandler struct {
	next ContextHandler
}

func (mp *OnTextHandler) Execute(m *Context) {
	slog.Debug("Entering OnTextHandler")
	switch m.Service {
	case Telegram:
		c := m.TelebotContext
		message := c.Message()
		if message != nil {
			m.rawText = strings.Replace(c.Message().Text, "@"+m.Telebot.Me.Username, "", 1)
			m.id = strconv.Itoa(c.Message().ID)
			m.isReply = c.Message().IsReply()
			m.chatId = strconv.Itoa(int(c.Chat().ID))

			if c.Message().IsReply() {
				m.replyToId = fmt.Sprint(c.Message().ReplyTo.ID)
				m.shouldReplyToMessage = true
			}
		}
	case Discord:
		message := m.DiscordMessage
		if message != nil {
			m.rawText = message.Content
			m.id = message.ID
			if message.ReferencedMessage != nil {
				m.replyToId = message.ReferencedMessage.ID
				m.shouldReplyToMessage = true
			}
			m.chatId = message.ChannelID
		}
	case Matrix:
		if m.MatrixEvent != nil && *m.MatrixEvent != nil {
			evt := (*m.MatrixEvent).(*event.Event)
			content := evt.Content.AsMessage()
			if content != nil {
				m.rawText = content.Body
			}
			m.id = evt.ID.String()
			m.chatId = evt.RoomID.String()
			
			if relatesTo := content.GetRelatesTo(); relatesTo != nil && relatesTo.GetReplyTo() != "" {
				m.replyToId = relatesTo.GetReplyTo().String()
				m.shouldReplyToMessage = true
			}
		}
	}
	mp.next.Execute(m)
}

func (mp *OnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
