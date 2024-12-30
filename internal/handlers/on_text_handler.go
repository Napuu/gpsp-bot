package handlers

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type OnTextHandler struct {
	next ContextHandler
}

func (mp *OnTextHandler) Execute(m *Context) {
	slog.Debug("Entering OnTextHandler")
	cfg := config.FromEnv()
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
				m.replyToId = string(rune(c.Message().ReplyTo.ID))
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
		message := m.MatrixEvent
		isOldMessage := int(time.Now().UnixMilli()-message.Timestamp) > utils.S2I(cfg.MATRIX_OLD_MESSAGE_THRESHOLD_MILLIS)
		if message != nil && !isOldMessage {
			m.rawText = message.Content.AsMessage().Body
			m.id = string(message.ID)
			m.chatId = string(message.RoomID)
			if message.Content.AsMessage().RelatesTo != nil {
				m.replyToId = message.Content.AsMessage().RelatesTo.InReplyTo.EventID.String()
				m.shouldReplyToMessage = true
			}
		}
	}
	mp.next.Execute(m)
}

func (mp *OnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
