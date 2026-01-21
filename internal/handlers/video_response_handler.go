package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type VideoResponseHandler struct {
	next ContextHandler
}

func (r *VideoResponseHandler) Execute(m *Context) {
	slog.Debug("Entering VideoResponseHandler")

	if len(m.finalVideoPath) > 0 {
		switch m.Service {
		case Telegram:
			chatId := tele.ChatID(utils.S2I(m.chatId))

			var sentMessage *tele.Message
			var err error
			if m.shouldReplyToMessage {
				message := &tele.Message{
					Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
					ID:   utils.S2I(m.replyToId),
				}
				sentMessage, err = m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.finalVideoPath)}, &tele.SendOptions{ReplyTo: message})
			} else {
				sentMessage, err = m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.finalVideoPath)})
			}
			if err != nil {
				slog.Warn("Failed to send video", "error", err)
			} else {
				m.sendVideoSucceeded = true
				// Store fingerprint with the message ID of the bot's response
				if len(m.pendingFingerprint) > 0 {
					messageId := fmt.Sprint(sentMessage.ID)
					if err := utils.StoreFingerprint(m.pendingFingerprintDbPath, m.pendingFingerprint, m.pendingFingerprintGroupId, messageId); err != nil {
						slog.Warn("Failed to store fingerprint", "error", err)
					} else {
						slog.Debug("Fingerprint stored", "groupId", m.pendingFingerprintGroupId, "messageId", messageId)
					}
				}
			}
		case Discord:
			file, err := os.Open(m.finalVideoPath)
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
				Content: "",
				Files: []*discordgo.File{
					{
						Name:        "video.mp4", // this apparently doesn't matter
						ContentType: "video/mp4",
						Reader:      buf,
					},
				},
			}

			if m.shouldReplyToMessage {
				message.Reference = &discordgo.MessageReference{
					ChannelID: m.chatId,
					MessageID: m.replyToId,
				}
			}

			sentMessage, err := m.DiscordSession.ChannelMessageSendComplex(m.chatId, message)
			if err != nil {
				slog.Warn("Failed to send video", "error", err)
			} else {
				m.sendVideoSucceeded = true
				// Store fingerprint with the message ID of the bot's response
				if len(m.pendingFingerprint) > 0 {
					if err := utils.StoreFingerprint(m.pendingFingerprintDbPath, m.pendingFingerprint, m.pendingFingerprintGroupId, sentMessage.ID); err != nil {
						slog.Warn("Failed to store fingerprint", "error", err)
					} else {
						slog.Debug("Fingerprint stored", "groupId", m.pendingFingerprintGroupId, "messageId", sentMessage.ID)
					}
				}
			}
		}
	}

	r.next.Execute(m)
}

func (u *VideoResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
