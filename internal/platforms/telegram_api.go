package platforms

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
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
	bot := getTelegramBot(dbPath)
	chain := chain.NewChainOfResponsibility()

	err := bot.SetCommands(TelebotCompatibleVisibleCommands())
	if err != nil {
		slog.Error(err.Error())
	}

	bot.Handle(tele.OnText, wrapTeleHandler(bot, chain))

	go bot.Start()
}

func getTelegramBot(dbPath string) *tele.Bot {
	inner := &tele.LongPoller{
		Timeout: 10 * time.Second,
		AllowedUpdates: []string{
			"message",
			"message_reaction",
		},
	}

	poller := tele.NewMiddlewarePoller(inner, func(u *tele.Update) bool {
		if u.MessageReaction != nil {
			mr := u.MessageReaction
			groupId := "telegram:" + fmt.Sprint(mr.Chat.ID)
			botMsgId := fmt.Sprint(mr.MessageID)

			// Find added reactions (present in New but not Old)
			for _, r := range mr.NewReaction {
				if !containsEmoji(mr.OldReaction, r.Emoji) {
					if err := utils.UpdateReactionCount(dbPath, "telegram", groupId, botMsgId, r.Emoji, +1); err != nil {
						slog.Warn("Failed to update Telegram reaction count", "error", err)
					}
				}
			}
			// Find removed reactions (present in Old but not New)
			for _, r := range mr.OldReaction {
				if !containsEmoji(mr.NewReaction, r.Emoji) {
					if err := utils.UpdateReactionCount(dbPath, "telegram", groupId, botMsgId, r.Emoji, -1); err != nil {
						slog.Warn("Failed to update Telegram reaction count", "error", err)
					}
				}
			}
		}
		return true
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

// containsEmoji reports whether emoji e is present in the reaction slice.
func containsEmoji(reactions []tele.Reaction, emoji string) bool {
	for _, r := range reactions {
		if r.Emoji == emoji {
			return true
		}
	}
	return false
}
