package main

import (
	"log"
	"os"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v4"
)

type Recipient struct {
	Recipient string
}

type TelegramChatId int64

func (i TelegramChatId) Recipient() string {
	return strconv.FormatInt(int64(i), 10)
}

func runTelegramBot() {
    bot := getTelegramBot()

    chain := NewChainOfResponsibility()

    bot.Handle(tele.OnText, func(c tele.Context) error {
	    chain.Process(&Context{telebotContext: c, telebot: bot, service: Telegram})
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
