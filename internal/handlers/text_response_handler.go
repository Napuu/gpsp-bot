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

	// Skip sending text if an image is being sent (image handler will send it with caption)
	if len(m.finalImagePath) > 0 {
		r.next.Execute(m)
		return
	}

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
		opts := &tele.SendOptions{DisableWebPagePreview: m.disableWebPreview}
		if m.shouldReplyToMessage {
			chatId := tele.ChatID(utils.S2I(m.chatId))
			opts.ReplyTo = &tele.Message{
				Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
				ID:   utils.S2I(m.replyToId),
			}
			m.Telebot.Send(chatId, m.textResponse, opts)
		} else if m.textResponse != "" {
			m.TelebotContext.Send(m.textResponse, opts)
		}

	case Discord:
		if m.textResponse == "" {
			break
		}
		send := &discordgo.MessageSend{Content: m.textResponse}
		if m.disableWebPreview {
			send.Flags = discordgo.MessageFlagsSuppressEmbeds
		}
		if m.shouldReplyToMessage {
			send.Reference = &discordgo.MessageReference{
				ChannelID: m.chatId,
				MessageID: m.replyToId,
			}
		}
		m.DiscordSession.ChannelMessageSendComplex(m.chatId, send)
	}

	r.next.Execute(m)
}

func (u *TextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
