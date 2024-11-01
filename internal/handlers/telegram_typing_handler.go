package handlers

import (
	"log"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TelegramTypingHandler struct {
	next ContextHandler
}

func (t *TelegramTypingHandler) Execute(m *Context) {
    log.Println("Entering TelegramTypingHandler")
    if m.Service == Telegram && (m.action == DownloadVideo || m.action == SearchVideo) {
        m.doneTyping = make(chan struct{})

        go func() {
            ticker := time.NewTicker(4 * time.Second)
            defer ticker.Stop()

            _ = m.TelebotContext.Notify(tele.UploadingVideo)

            for {
                select {
                case <-m.doneTyping:
                    return
                case <-ticker.C:
                                log.Println("Continue typing")
                    _ = m.TelebotContext.Notify(tele.UploadingVideo)
                }
            }
        }()
    }

    t.next.Execute(m)
}


func (t *TelegramTypingHandler) SetNext(next ContextHandler) {
	t.next = next
}
