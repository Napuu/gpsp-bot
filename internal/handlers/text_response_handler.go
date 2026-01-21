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

	// Check for repost and set up reply to original message
	if m.isRepost && len(m.repostOriginalMessageIds) > 0 {
		warningMsg := "⚠️ Repost detected (similar to previously seen video)"
		if m.textResponse != "" {
			m.textResponse = warningMsg + "\n\n" + m.textResponse
		} else {
			m.textResponse = warningMsg
		}
		// Reply to the original message that first posted this video
		m.shouldReplyToMessage = true
		m.replyToId = m.repostOriginalMessageIds[0]
	}

	switch m.Service {
	case Telegram:
		if m.shouldReplyToMessage {
			chatId := tele.ChatID(utils.S2I(m.chatId))
			message := &tele.Message{
				Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
				ID:   utils.S2I(m.replyToId),
			}
			m.Telebot.Send(chatId, m.textResponse, &tele.SendOptions{ReplyTo: message})
		} else if m.textResponse != "" {
			m.TelebotContext.Send(m.textResponse)
		}

	case Discord:
		if m.shouldReplyToMessage {
			message := &discordgo.MessageReference{
				ChannelID: m.chatId,
				MessageID: m.replyToId,
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
