package platforms

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"

	tele "gopkg.in/telebot.v4"
)

func wrapTeleHandler(bot *tele.Bot, chain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		var replyToId string
		isReply := c.Message().ReplyTo != nil
		if isReply {
			replyToId = strconv.Itoa(c.Message().ReplyTo.ID)
		}

		chain.Process(&handlers.Context{
			TelebotContext:      c,
			Telebot:             bot,
			Service:             handlers.Telegram,
			RawText:             c.Message().Text,
			ID:                  strconv.Itoa(c.Message().ID),
			ChatID:              strconv.FormatInt(c.Message().Chat.ID, 10),
			IsReply:             isReply,
			ReplyToID:           replyToId,
			ShouldReplyToMessage: isReply,
		})
		return nil
	}
}

func TelebotCompatibleVisibleCommands() []tele.Command {
	commands := make([]tele.Command, 0, len(config.EnabledFeatures()))
	for _, action := range config.EnabledFeatures() {
		if handlers.Action(action) == handlers.Ping {
			continue
		}
		commands = append(commands, tele.Command{
			Text:        string(action),
			Description: string(handlers.ActionMap[handlers.Action(action)]),
		})
	}
	return commands
}

func RunTelegramBot() {
	bot := getTelegramBot()
	chain := chain.NewChainOfResponsibility()

	err := bot.SetCommands(TelebotCompatibleVisibleCommands())
	if err != nil {
		slog.Error(err.Error())
	}

	bot.Handle(tele.OnText, wrapTeleHandler(bot, chain))

	go bot.Start()
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
