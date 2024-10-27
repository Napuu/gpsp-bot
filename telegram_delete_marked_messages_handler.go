package main

import (
	"log"
)

type TelegramDeleteMarkedMessageHandler struct {
	next ContextHandler
}

func (r *TelegramDeleteMarkedMessageHandler) execute(m *Context) {
	log.Println("Entering TelegramDeleteMarkedMessageHandler")
	if (m.service == Telegram && m.shouldDeleteOriginalMessage) {
		err := m.telebotContext.Delete()
		if err != nil {
			log.Println(err)
		}
	}

	r.next.execute(m)
}

func (u *TelegramDeleteMarkedMessageHandler) setNext(next ContextHandler) {
	u.next = next
}