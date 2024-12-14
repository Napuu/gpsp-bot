package handlers

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type TextResponseHandler struct {
	next ContextHandler
}

func (r *TextResponseHandler) Execute(m *Context) {
	slog.Debug("Entering TextResponseHandler")
	switch m.Service {
	case Telegram:
		if m.shouldReplyToMessage {
			message := &tele.Message{
				Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
				ID:   utils.S2I(m.replyToId),
			}
			m.Telebot.Reply(message, m.textResponse)
		} else if m.textResponse != "" {
			m.TelebotContext.Send(m.textResponse)
		}

	case Discord:
		if m.shouldReplyToMessage {
			message := &discordgo.MessageReference{
				ChannelID: m.chatId,
				MessageID: m.id,
			}
			m.DiscordSession.ChannelMessageSendReply(m.chatId, m.textResponse, message)
		} else if m.textResponse != "" {
			m.DiscordSession.ChannelMessageSend(m.chatId, m.textResponse)
		}
	}

	r.next.Execute(m)
}

func (u *TextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
