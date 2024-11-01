package platforms

import (
	"log"
	"os"
	"time"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/handlers"

	tele "gopkg.in/telebot.v4"
)

func wrapHandler(bot *tele.Bot, chain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		chain.Process(&handlers.Context{TelebotContext: c, Telebot: bot, Service: handlers.Telegram})
		return nil
	}
}

func RunTelegramBot() {
    bot := getTelegramBot()
    chain := chain.NewChainOfResponsibility()

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
