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
		}
	}

	r.next.Execute(m)
}

func (u *ImageResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
