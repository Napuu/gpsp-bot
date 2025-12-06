package handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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
	
	case Matrix:
		if m.textResponse != "" && m.MatrixClient != nil && *m.MatrixClient != nil {
			client := (*m.MatrixClient).(*mautrix.Client)
			content := &event.MessageEventContent{
				MsgType: event.MsgText,
				Body:    m.textResponse,
			}
			
			if m.shouldReplyToMessage {
				content.RelatesTo = &event.RelatesTo{
					InReplyTo: &event.InReplyTo{
						EventID: id.EventID(m.replyToId),
					},
				}
			}
			
			_, err := client.SendMessageEvent(context.Background(), id.RoomID(m.chatId), event.EventMessage, content)
			if err != nil {
				slog.Error("Failed to send Matrix message", "error", err)
			}
		}
	}

	r.next.Execute(m)
}

func (u *TextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
