package main

import (
	"log"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TelegramTypingHandler struct {
	next ContextHandler
}

func (t *TelegramTypingHandler) execute(m *Context) {
    log.Println("Entering TelegramTypingHandler")
    if m.service == Telegram && (m.action == DownloadVideo || m.action == SearchVideo) {
        m.doneTyping = make(chan struct{})

        go func() {
            ticker := time.NewTicker(4 * time.Second)
            defer ticker.Stop()

            _ = m.telebotContext.Notify(tele.UploadingVideo)

            for {
                select {
                case <-m.doneTyping:
                    return
                case <-ticker.C:
                                log.Println("Continue typing")
                    _ = m.telebotContext.Notify(tele.UploadingVideo)
                }
            }
        }()
    }

    t.next.execute(m)
}


func (t *TelegramTypingHandler) setNext(next ContextHandler) {
	t.next = next
}
