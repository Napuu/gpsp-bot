package handlers

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type DiscordTextResponseHandler struct {
	next ContextHandler
}

func (r *DiscordTextResponseHandler) Execute(m *Context) {
	slog.Debug("Entering DiscordTextResponseHandler")
	if m.Service == Discord {
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

func (u *DiscordTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
