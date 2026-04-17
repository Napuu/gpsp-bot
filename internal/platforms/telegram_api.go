package platforms

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"github.com/napuu/gpsp-bot/internal/telereactions"
	"github.com/napuu/gpsp-bot/pkg/utils"

	tele "gopkg.in/telebot.v4"
)

func wrapTeleHandler(bot *tele.Bot, chain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		chain.Process(&handlers.Context{TelebotContext: c, Telebot: bot, Service: handlers.Telegram})
		return nil
	}
}

func TelebotCompatibleVisibleCommands() []tele.Command {
	commands := make([]tele.Command, 0, len(config.EnabledFeatures()))
	for _, action := range config.EnabledFeatures() {
		if handlers.Action(action) == handlers.Ping || handlers.Action(action) == handlers.Version {
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
	cfg := config.FromEnv()
	dbPath := filepath.Join(cfg.REPOST_DB_DIR, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		slog.Error("Failed to initialize stats DB", "error", err)
		return
	}
	bot := getTelegramBot(dbPath)
	chain := chain.NewChainOfResponsibility()

	if err := bot.SetCommands(TelebotCompatibleVisibleCommands()); err != nil {
		slog.Error(err.Error())
	}

	bot.Handle(tele.OnText, wrapTeleHandler(bot, chain))

	go bot.Start()
}

func getTelegramBot(dbPath string) *tele.Bot {
	inner := &tele.LongPoller{
		Timeout:        10 * time.Second,
		AllowedUpdates: []string{"message"},
	}

	updateCount := func(e telereactions.Event, delta int) {
		db, err := utils.OpenStatsDB(dbPath)
		if err != nil {
			slog.Warn("Failed to open stats DB for reaction", "error", err)
			return
		}
		defer db.Close()
		groupId := "telegram:" + fmt.Sprint(e.Chat.ID)
		if err := utils.UpdateReactionCount(db, "telegram", groupId, fmt.Sprint(e.MessageID), e.Emoji, delta); err != nil {
			slog.Warn("Failed to update Telegram reaction count", "error", err)
		}
	}
	poller := telereactions.Wrap(inner, telereactions.Handlers{
		OnAdd:    func(e telereactions.Event) { updateCount(e, +1) },
		OnRemove: func(e telereactions.Event) { updateCount(e, -1) },
	})

	pref := tele.Settings{
		Token:     config.FromEnv().TELEGRAM_TOKEN,
		ParseMode: tele.ModeHTML,
		Poller:    poller,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		panic(err)
	}

	return b
}
