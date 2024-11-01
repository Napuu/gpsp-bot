package main

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v4"
)

func wrapHandler(bot *tele.Bot, chain *HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		chain.Process(&Context{telebotContext: c, telebot: bot, service: Telegram})
		return nil
	}
}

func runTelegramBot() {
    bot := getTelegramBot()
    chain := NewChainOfResponsibility()

		bot.Handle(tele.OnMessageReaction, wrapHandler(bot, chain))
		bot.Handle(tele.OnText, wrapHandler(bot, chain))

    log.Println("Starting Telegram bot...")
    bot.Start()
}


func getTelegramBot() *tele.Bot {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
			AllowedUpdates: []string{
				"message",
				// TODO - take this into use
				// "message_reaction",
			},
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Panic(err)
	}

	return b
}
