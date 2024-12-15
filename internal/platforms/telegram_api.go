package platforms

import (
	"time"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"golang.org/x/exp/slog"

	tele "gopkg.in/telebot.v4"
)

func wrapTeleHandler(bot *tele.Bot, chain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		chain.Process(&handlers.Context{TelebotContext: c, Telebot: bot, Service: handlers.Telegram})
		return nil
	}
}

func TelebotCompatibleVisibleCommands() []tele.Command {
	visible := handlers.VisibleCommands()
	commands := make([]tele.Command, 0, len(visible))
	for action, description := range visible {
		commands = append(commands, tele.Command{
			Text:        string(action),
			Description: string(description),
		})
	}
	return commands
}

func RunTelegramBot() {
	bot := getTelegramBot()
	chain := chain.NewChainOfResponsibility()

	bot.SetCommands(TelebotCompatibleVisibleCommands())

	// bot.Handle(tele.OnMessageReaction, wrapHandler(bot, chain))
	bot.Handle(tele.OnText, wrapTeleHandler(bot, chain))

	slog.Info("Starting Telegram bot...")
	bot.Start()
}

func getTelegramBot() *tele.Bot {
	pref := tele.Settings{
		Token:     config.FromEnv().TELEGRAM_TOKEN,
		ParseMode: tele.ModeHTML,
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
		panic(err)
	}

	return b
}
