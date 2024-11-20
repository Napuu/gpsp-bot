package handlers

import (
	"log/slog"
	"time"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

type TelegramTuplillaResponseHandler struct {
	next ContextHandler
}

func (r *TelegramTuplillaResponseHandler) Execute(m *Context) {
	slog.Debug("Entering TelegramTuplillaResponseHandler")
	if m.Service == Telegram && m.action == Tuplilla {
		chatId := tele.ChatID(utils.S2I(m.chatId))

		cube1Response, err := m.Telebot.Send(chatId, tele.Cube)

		if err != nil {
			slog.Error(err.Error())
		}

		time.Sleep(2 * time.Second)

		cube2Response, err := m.Telebot.Send(chatId, tele.Cube)

		if err != nil {
			slog.Error(err.Error())
		}

		if cube1Response.Dice.Value == cube2Response.Dice.Value {
			m.gotDubz = true
		}

		m.lastCubeThrownTime = time.Now()
		m.dubzNegation = make(chan string)
		go func() {
			m.dubzNegation <- utils.GetNegation(m.parsedText)
		}()
	}

	r.next.Execute(m)
}

func (u *TelegramTuplillaResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
