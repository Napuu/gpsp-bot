package main

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v4"
)

func runTelegramBot() {
    bot := getTelegramBot()

    chain := NewHandlerChain()
    
    bot.Handle(tele.OnText, func(c tele.Context) error {
			chain.Process(&GenericMessage{telebotContext: c, service: Telegram})
			return nil
    })

    log.Println("Starting Telegram bot...")
    bot.Start()
}


func getTelegramBot() *tele.Bot {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Panic(err)
	}

	return b
}
