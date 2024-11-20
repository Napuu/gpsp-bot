package handlers

import (
	"log/slog"
	"time"
)

type DiscordTypingHandler struct {
	next ContextHandler
}

func (t *DiscordTypingHandler) Execute(m *Context) {
	slog.Debug("Entering DiscordTypingHandler")
	if m.Service == Discord && m.action != "" {
		m.doneTyping = make(chan struct{})

		go func() {
			ticker := time.NewTicker(4 * time.Second)
			defer ticker.Stop()

			_ = m.DiscordSession.ChannelTyping(m.chatId)

			for {
				select {
				case <-m.doneTyping:
					return
				case <-ticker.C:
					slog.Debug("Continue typing")
					_ = m.DiscordSession.ChannelTyping(m.chatId)
				}
			}
		}()
	}

	t.next.Execute(m)
}

func (t *DiscordTypingHandler) SetNext(next ContextHandler) {
	t.next = next
}
