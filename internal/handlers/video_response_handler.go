package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	tele "gopkg.in/telebot.v4"
)

type VideoResponseHandler struct {
	next ContextHandler
}

type ffprobeStream struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
}

func getVideoDimensions(path string) (width int, height int) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", path)
	output, err := cmd.Output()
	if err != nil {
		slog.Debug("Failed to get video dimensions with ffprobe", "error", err)
		return 0, 0
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		slog.Debug("Failed to parse ffprobe output", "error", err)
		return 0, 0
	}

	if len(probe.Streams) > 0 {
		return probe.Streams[0].Width, probe.Streams[0].Height
	}

	return 0, 0
}

func (r *VideoResponseHandler) Execute(m *Context) {
	slog.Debug("Entering VideoResponseHandler")

	if len(m.finalVideoPath) > 0 {
		switch m.Service {
		case Telegram:
			chatId := tele.ChatID(utils.S2I(m.chatId))

			if m.shouldReplyToMessage {
				message := &tele.Message{
					Chat: &tele.Chat{ID: int64(utils.S2I(m.chatId))},
					ID:   utils.S2I(m.replyToId),
				}
				m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.finalVideoPath)}, &tele.SendOptions{ReplyTo: message})
			} else {
				m.Telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.finalVideoPath)})
			}
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

			if m.shouldReplyToMessage {
				message.Reference = &discordgo.MessageReference{
					ChannelID: m.chatId,
					MessageID: m.id,
				}
			}

			_, err = m.DiscordSession.ChannelMessageSendComplex(m.chatId, message)
			if err != nil {
				slog.Debug(err.Error())
			} else {
				m.sendVideoSucceeded = true
			}
		case Matrix:
			if m.MatrixClient != nil && *m.MatrixClient != nil {
				client := (*m.MatrixClient).(*mautrix.Client)
				
				data, err := os.ReadFile(m.finalVideoPath)
				if err != nil {
					slog.Error("Failed to read video file", "error", err)
					break
				}

				uploaded, err := client.UploadMedia(context.Background(), mautrix.ReqUploadMedia{
					Content:       bytes.NewReader(data),
					ContentLength: int64(len(data)),
					ContentType:   "video/mp4",
				})
				if err != nil {
					slog.Error("Failed to upload video to Matrix", "error", err)
					break
				}

				// Get video dimensions for proper rendering
				width, height := getVideoDimensions(m.finalVideoPath)
				
				videoInfo := &event.FileInfo{
					MimeType: "video/mp4",
					Size:     len(data),
				}
				
				if width > 0 && height > 0 {
					videoInfo.Width = width
					videoInfo.Height = height
					slog.Debug("Video dimensions", "width", width, "height", height)
				}

				content := &event.MessageEventContent{
					MsgType: event.MsgVideo,
					Body:    "video.mp4",
					URL:     uploaded.ContentURI.CUString(),
					Info:    videoInfo,
				}

				if m.shouldReplyToMessage {
					content.RelatesTo = &event.RelatesTo{
						InReplyTo: &event.InReplyTo{
							EventID: id.EventID(m.replyToId),
						},
					}
				}

				_, err = client.SendMessageEvent(context.Background(), id.RoomID(m.chatId), event.EventMessage, content)
				if err != nil {
					slog.Error("Failed to send video message", "error", err)
				} else {
					m.sendVideoSucceeded = true
				}
			}
		}
	}

	r.next.Execute(m)
}

func (u *VideoResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
