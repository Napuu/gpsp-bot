package main

import (
	"log"

	tele "gopkg.in/telebot.v4"
)

type TelegramVideoResponseHandler struct {
	next ContextHandler
}


func (r *TelegramVideoResponseHandler) execute(m *Context) {
	log.Println("Entering TelegramVideoResponseHandler")
	if (m.service == Telegram) {
		chatId := tele.ChatID(m.chatId)

		var videoPathToSend = m.originalVideoPath
		if len(m.possiblyProcessedVideoPath) > 0 {
			videoPathToSend = m.possiblyProcessedVideoPath
		}

		if len(videoPathToSend) > 0 {
			m.telebot.Send(chatId, &tele.Video{File: tele.FromDisk(videoPathToSend)} )
			m.sendVideoSucceeded = true
		}
	}

	r.next.execute(m)
}

func (u *TelegramVideoResponseHandler) setNext(next ContextHandler) {
	u.next = next
}
