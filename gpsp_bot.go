package main

import (
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

func (telegramMessageParser *TelegramMessageParser) execute(c tele.Context) {
	genericMessage := GenericMessage{
		text: c.Message().Text,
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

// func (handler )

func main() {
	// pref := tele.Settings{
	// 	Token:  os.Getenv("TOKEN"),
	// 	Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	// }

	// b, err := tele.NewBot(pref)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	// genericMessageParser := &GenericMessageParser{}
	// telegramMessageParser := &TelegramMessageParser{}
	// telegramMessageParser.setNext(genericMessageParser)

	// b.Handle(tele.OnText, func(c tele.Context) error {
	// 	telegramMessageParser.execute(c)
	// 	return nil
	// })

	// b.Handle("*", func(c tele.Context) error {
		
	// 	return c.Send("Hello!")
	// })

	// b.Start()

	urlParser := &URLParser{}
	genericMessageParser := &GenericMessageParser{}
	genericMessageParser.setNext(urlParser)
	telegramMessageParser := &TelegramMessageParser{}
	telegramMessageParser.setNext(genericMessageParser)
	// foo := []string
	// foo := string{"123", "foo"}
	genericMessageParser.execute(&GenericMessage{
		text: "/tuplilla foobar",
		id: 123,
		isReply: true,
		replyToId: 456,
		chatId: 123456,
	})
}