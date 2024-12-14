package handlers

import (
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TypingHandler struct {
	next ContextHandler
}

func sendTyping(m *Context) {
	var err error

	switch m.Service {
	case Telegram:
		action := tele.Typing
		if m.action == DownloadVideo || m.action == SearchVideo {
			action = tele.UploadingVideo
		}
		err = m.TelebotContext.Notify(action)
	case Discord:
		err = m.DiscordSession.ChannelTyping(m.chatId)
	}

	if err != nil {
		slog.Error(err.Error())
	}
}

func (t *TypingHandler) Execute(m *Context) {
	slog.Debug("Entering TypingHandler")
	if m.action != "" {
		m.doneTyping = make(chan struct{})

		go func() {
			ticker := time.NewTicker(4 * time.Second)
			defer ticker.Stop()

			sendTyping(m)

			for {
				select {
				case <-m.doneTyping:
					return
				case <-ticker.C:
					slog.Debug("Continue typing")
					sendTyping(m)
				}
			}
		}()
	}

	t.next.Execute(m)
}

func (t *TypingHandler) SetNext(next ContextHandler) {
	t.next = next
}
