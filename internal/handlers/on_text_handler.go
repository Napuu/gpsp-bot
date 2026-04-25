package handlers

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
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
			m.chatUsername = c.Chat().Username

			if c.Message().IsReply() {
				m.replyToId = fmt.Sprint(c.Message().ReplyTo.ID)
				m.shouldReplyToMessage = true
			}

			if message.Sender != nil {
				m.posterUserId = strconv.FormatInt(message.Sender.ID, 10)
				m.posterUsername = message.Sender.Username
				if m.posterUsername == "" {
					m.posterUsername = strings.TrimSpace(message.Sender.FirstName + " " + message.Sender.LastName)
				}
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
			m.guildId = message.GuildID

			if message.Author != nil {
				m.posterUserId = message.Author.ID
				m.posterUsername = message.Author.Username
			}
		}
	}
	mp.next.Execute(m)
}

func (mp *OnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
