package main

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v4"
)

type handler interface {
	execute(*GenericMessage)
	setNext(handler)
}

type responder interface {

}

type TelegramMessageParser struct {
	next handler
}

func (telegramMessageParser *TelegramMessageParser) execute(m *GenericMessage) {
	c := m.telebotContext
	genericMessage := GenericMessage{
		rawText: c.Message().Text,
		id: c.Message().ID,
		isReply: c.Message().IsReply(),
		chatId: c.Chat().ID,
	}

	if (c.Message().IsReply()) {
		genericMessage.replyToId = c.Message().ReplyTo.ID
	}

	telegramMessageParser.next.execute(&genericMessage)
}

func (mp *TelegramMessageParser) setNext(next handler) {
	mp.next = next
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	urlParser := HandlerLogger(&URLParser{})
	genericMessageParser := HandlerLogger(&GenericMessageParser{})
	genericMessageParser.setNext(urlParser)
	telegramMessageParser := HandlerLogger(&TelegramMessageParser{})
	telegramMessageParser.setNext(genericMessageParser)
	b.Handle(tele.OnText, func(c tele.Context) error {
		log.Println("receive message?")

		telegramMessageParser.execute(&GenericMessage{telebotContext: c})
		return nil
	})

	// b.Handle("*", func(c tele.Context) error {
		
	// 	return c.Send("Hello!")
	// })

	b.Start()

	// foo := []string
	// foo := string{"123", "foo"}
	// genericMessageParser.execute(&GenericMessage{
	// 	// text: "/tuplilla foobar",
	// 	rawText: "/dl https://youtube.com/asdfwera?w=234kissa",
	// 	id: 123,
	// 	isReply: true,
	// 	replyToId: 456,
	// 	chatId: 123456,
	// })
}