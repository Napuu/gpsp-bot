package main

import (
	"log"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TelegramTuplillaResponseHandler struct {
	next ContextHandler
}

func (r *TelegramTuplillaResponseHandler) execute(m *Context) {
	log.Println("Entering TelegramTuplillaResponseHandler")
	if m.service == Telegram && m.action == Tuplilla {
		chatId := tele.ChatID(m.chatId)

		cube1Response, err := m.telebot.Send(chatId, tele.Cube)

		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(2 * time.Second)

		cube2Response, err := m.telebot.Send(chatId, tele.Cube)

		if err != nil {
			log.Fatal(err)
		}

		if cube1Response.Dice.Value == cube2Response.Dice.Value {
			m.gotDubz = true
		}

		m.lastCubeThrownTime = time.Now()
	}

	r.next.execute(m)
}

func (u *TelegramTuplillaResponseHandler) setNext(next ContextHandler) {
	u.next = next
}