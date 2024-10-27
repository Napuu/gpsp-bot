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

		if (m.action == DownloadVideo && len(m.downloadedVideoPath) > 0) {
			m.telebot.Send(chatId, &tele.Video{File: tele.FromDisk(m.downloadedVideoPath)} )
			m.sendVideoSucceeded = true
		}
	}

	r.next.execute(m)
}

func (u *TelegramVideoResponseHandler) setNext(next ContextHandler) {
	u.next = next
}
