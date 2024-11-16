package handlers

import (
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TelegramTypingHandler struct {
	next ContextHandler
}

func (t *TelegramTypingHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramTypingHandler")
	if m.Service == Telegram && m.action != "" {
		action := tele.Typing
		if m.action == DownloadVideo || m.action == SearchVideo {
			action = tele.UploadingVideo
		}
		m.doneTyping = make(chan struct{})

		go func() {
			ticker := time.NewTicker(4 * time.Second)
			defer ticker.Stop()

			_ = m.TelebotContext.Notify(action)

			for {
				select {
				case <-m.doneTyping:
					return
				case <-ticker.C:
					slog.Debug("Continue typing")
					_ = m.TelebotContext.Notify(action)
				}
			}
		}()
	}

	t.next.Execute(m)
}

func (t *TelegramTypingHandler) SetNext(next ContextHandler) {
	t.next = next
}
