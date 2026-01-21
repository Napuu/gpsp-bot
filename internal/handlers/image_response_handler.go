package handlers

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type ImageResponseHandler struct {
	next ContextHandler
}

func (r *ImageResponseHandler) Execute(m *Context) {
	slog.Debug("Entering ImageResponseHandler")

	if len(m.finalImagePath) > 0 {
		switch m.Service {
		case Telegram:
			chatId := tele.ChatID(utils.S2I(m.chatId))

			photo := &tele.Photo{File: tele.FromDisk(m.finalImagePath)}
			if m.textResponse != "" {
				photo.Caption = m.textResponse
			}

			var sentMessage *tele.Message
			var err error
			// Check image-specific reply fields first (used for repost images), fallback to shared fields
			shouldReply := m.imageShouldReplyToMessage || m.shouldReplyToMessage
			replyToId := m.imageReplyToId
			if replyToId == "" {
				replyToId = m.replyToId
			}
			if shouldReply {
				message := &tele.Message{
					Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
					ID:   utils.S2I(replyToId),
				}
				sentMessage, err = m.Telebot.Send(chatId, photo, &tele.SendOptions{ReplyTo: message})
			} else {
				sentMessage, err = m.Telebot.Send(chatId, photo)
			}
			if err != nil {
				slog.Warn("Failed to send image", "error", err)
			} else {
				slog.Debug("Image sent successfully", "messageId", sentMessage.ID)
			}
		case Discord:
			file, err := os.Open(m.finalImagePath)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			buf := bytes.NewBuffer(nil)
			_, err = buf.ReadFrom(file)
			if err != nil {
				panic(err)
			}

			message := &discordgo.MessageSend{
				Content: m.textResponse,
				Files: []*discordgo.File{
					{
						Name:        "image.jpg", // this apparently doesn't matter
						ContentType: "image/jpeg",
						Reader:      buf,
					},
				},
			}

			// Check image-specific reply fields first (used for repost images), fallback to shared fields
			shouldReply := m.imageShouldReplyToMessage || m.shouldReplyToMessage
			replyToId := m.imageReplyToId
			if replyToId == "" {
				replyToId = m.replyToId
			}
			if shouldReply {
				message.Reference = &discordgo.MessageReference{
					ChannelID: m.chatId,
					MessageID: replyToId,
				}
			}

			_, err = m.DiscordSession.ChannelMessageSendComplex(m.chatId, message)
			if err != nil {
				slog.Debug(err.Error())
			}
		}
	}

	r.next.Execute(m)
}

func (u *ImageResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
