package handlers

import (
	"bytes"
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

			m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.finalVideoPath)})
			m.sendVideoSucceeded = true
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

			_, err = m.DiscordSession.ChannelMessageSendComplex(m.chatId, message)
			if err != nil {
				slog.Debug(err.Error())
			} else {
				m.sendVideoSucceeded = true
			}
		}
	}

	r.next.Execute(m)
}

func (u *VideoResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
