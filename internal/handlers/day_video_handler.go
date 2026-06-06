package handlers

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/dayvideo"
	tele "gopkg.in/telebot.v4"
)

// DayVideoHandler records group activity and may trigger the day meme easter egg.
type DayVideoHandler struct {
	next      ContextHandler
	scheduler *dayvideo.Scheduler
}

// NewDayVideoHandler creates a handler wired to the day video scheduler.
func NewDayVideoHandler(scheduler *dayvideo.Scheduler) *DayVideoHandler {
	return &DayVideoHandler{scheduler: scheduler}
}

func (h *DayVideoHandler) Execute(m *Context) {
	if h.scheduler != nil {
		switch m.Service {
		case Telegram:
			h.handleTelegram(m)
		case Discord:
			h.handleDiscord(m)
		}
	}
	h.next.Execute(m)
}

func (h *DayVideoHandler) handleTelegram(m *Context) {
	c := m.TelebotContext
	if c == nil || c.Message() == nil || c.Chat() == nil {
		return
	}
	if c.Message().Sender != nil && m.Telebot != nil && c.Message().Sender.ID == m.Telebot.Me.ID {
		return
	}

	chat := c.Chat()
	if chat.Type != tele.ChatGroup && chat.Type != tele.ChatSuperGroup {
		return
	}

	memberCount := 0
	if m.Telebot != nil {
		if count, err := m.Telebot.Len(chat); err == nil {
			memberCount = count
		}
	}

	h.scheduler.RecordActivityAndTry("telegram", m.chatId, true, memberCount)
}

func (h *DayVideoHandler) handleDiscord(m *Context) {
	if m.DiscordMessage == nil || m.guildId == "" {
		return
	}

	memberCount := dayvideo.DiscordMemberCount(m.DiscordSession, m.guildId)
	if memberCount == 0 {
		slog.Debug("day video: discord member count unavailable", "guildId", m.guildId)
	}

	h.scheduler.RecordActivityAndTry("discord", m.chatId, true, memberCount)
}

func (h *DayVideoHandler) SetNext(next ContextHandler) {
	h.next = next
}
