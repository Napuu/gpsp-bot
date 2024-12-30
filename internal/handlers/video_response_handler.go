package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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
		case Matrix:
			// TODO - send using matrix-go
			roomId := m.chatId // Assuming chatId corresponds to Matrix's room ID
			file, err := os.Open(m.finalVideoPath)
			if err != nil {
				slog.Error("Failed to open video file", "error", err)
				return
			}
			defer file.Close()

			// uploadRequest := mautrix.ReqUploadMedia{
			// 	FileName: "video.mp4",
			// 	Content:  file,
			// }
			// uploadResponse, err := m.MatrixClient.UploadMedia(context.TODO(), uploadRequest)
			// if err != nil {
			// 	slog.Error("Failed to upload video to Matrix", "error", err)
			// 	return
			// }
			// // fmt.Println("rep", uploadResponse)
			// fmt.Printf("%+v\n", uploadResponse)

			// Send the video as a message
			videoMessage := map[string]interface{}{
				"msgtype": "m.video",
				"body":    "video.mp4",
				"url":     "mxc://matrix.napuu.fi/619315d83adda3a2064c7e61e1fc703c61cd6239519b70a6c4b2fe8278a56469",
				"info": map[string]interface{}{
					"mimetype": "video/mp4",
				},
			}

			_, err = m.MatrixClient.SendMessageEvent(context.TODO(), id.RoomID(roomId), event.EventMessage, videoMessage)
			if err != nil {
				slog.Error("Failed to send video message to Matrix", "error", err)
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
