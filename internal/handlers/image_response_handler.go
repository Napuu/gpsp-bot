package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

			m.Telebot.Send(chatId, &tele.Photo{File: tele.FromDisk(m.finalImagePath)})
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
				Content: "",
				Files: []*discordgo.File{
					{
						Name:        "image.jpg", // this apparently doesn't matter
						ContentType: "image/jpeg",
						Reader:      buf,
					},
				},
			}

			_, err = m.DiscordSession.ChannelMessageSendComplex(m.chatId, message)
			if err != nil {
				slog.Debug(err.Error())
			}
		case Matrix:
			if m.MatrixClient != nil && *m.MatrixClient != nil {
				client := (*m.MatrixClient).(*mautrix.Client)
				
				data, err := os.ReadFile(m.finalImagePath)
				if err != nil {
					slog.Error("Failed to read image file", "error", err)
					break
				}

				uploaded, err := client.UploadMedia(context.Background(), mautrix.ReqUploadMedia{
					Content:       bytes.NewReader(data),
					ContentLength: int64(len(data)),
					ContentType:   "image/jpeg",
				})
				if err != nil {
					slog.Error("Failed to upload image to Matrix", "error", err)
					break
				}

				content := &event.MessageEventContent{
					MsgType: event.MsgImage,
					Body:    "image.jpg",
					URL:     uploaded.ContentURI.CUString(),
					Info: &event.FileInfo{
						MimeType: "image/jpeg",
						Size:     len(data),
					},
				}

				_, err = client.SendMessageEvent(context.Background(), id.RoomID(m.chatId), event.EventMessage, content)
				if err != nil {
					slog.Error("Failed to send image message", "error", err)
				}
			}
		}
	}

	r.next.Execute(m)
}

func (u *ImageResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
