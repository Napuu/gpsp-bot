package dayvideo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

// PostContext holds platform details needed to send a day meme video.
type PostContext struct {
	GroupID        string
	Platform       string
	ChatID         string
	Telebot        *tele.Bot
	DiscordSession *discordgo.Session
}

// Poster sends generated day meme videos to a chat.
type Poster interface {
	PostDayVideo(ctx PostContext, videoPath string) error
}

type livePoster struct{}

func (livePoster) PostDayVideo(ctx PostContext, videoPath string) error {
	switch ctx.Platform {
	case "telegram":
		if ctx.Telebot == nil {
			return fmt.Errorf("telegram bot not configured")
		}
		chatID := tele.ChatID(utils.S2I(ctx.ChatID))
		_, err := ctx.Telebot.Send(chatID, &tele.Video{File: tele.FromDisk(videoPath)})
		if err != nil {
			return fmt.Errorf("send telegram day video: %w", err)
		}
		return nil
	case "discord":
		if ctx.DiscordSession == nil {
			return fmt.Errorf("discord session not configured")
		}
		file, err := os.Open(videoPath)
		if err != nil {
			return fmt.Errorf("open day video: %w", err)
		}
		defer file.Close()

		buf := bytes.NewBuffer(nil)
		if _, err := buf.ReadFrom(file); err != nil {
			return fmt.Errorf("read day video: %w", err)
		}

		_, err = ctx.DiscordSession.ChannelMessageSendComplex(ctx.ChatID, &discordgo.MessageSend{
			Files: []*discordgo.File{
				{
					Name:        "day.mp4",
					ContentType: "video/mp4",
					Reader:      buf,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("send discord day video: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform %q", ctx.Platform)
	}
}

// NewPoster returns the production poster implementation.
func NewPoster() Poster {
	return livePoster{}
}

// GenerateToTemp creates a day meme video for the given time in a temp file.
func GenerateToTemp(at time.Time) (string, error) {
	path := filepath.Join(utils.DayVideoTmpDir(), uuid.New().String()+".mp4")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := utils.GenerateDayVideo(path, at); err != nil {
		return "", err
	}
	return path, nil
}
