package platforms

import (
	"log/slog"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

func wrapDiscoHandler(chain *chain.HandlerChain) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Wrap the context
		chain.Process(&handlers.Context{
			DiscordSession: s,
			DiscordMessage: m,
			Service:        handlers.Discord,
		})
	}
}

func RunDiscordBot() {
	cfg := config.FromEnv()
	token := cfg.DISCORD_TOKEN
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("Error creating Discord session", "error", err)
		return
	}

	// Create the chain of responsibility
	chain := chain.NewChainOfResponsibility()

	dbPath := filepath.Join(cfg.REPOST_DB_DIR, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		slog.Error("Failed to initialize stats DB", "error", err)
		return
	}
	statsDB, err := utils.OpenStatsDB(dbPath)
	if err != nil {
		slog.Error("Failed to open stats DB for reaction tracking", "error", err)
		return
	}

	// Add a handler for messages
	dg.AddHandler(wrapDiscoHandler(chain))

	// Add reaction tracking handlers
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.UserID == s.State.User.ID {
			return
		}
		groupId := "discord:" + r.ChannelID
		if err := utils.UpdateReactionCount(statsDB, "discord", groupId, r.MessageID, r.Emoji.Name, +1); err != nil {
			slog.Warn("Failed to update Discord reaction count", "error", err)
		}
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if r.UserID == s.State.User.ID {
			return
		}
		groupId := "discord:" + r.ChannelID
		if err := utils.UpdateReactionCount(statsDB, "discord", groupId, r.MessageID, r.Emoji.Name, -1); err != nil {
			slog.Warn("Failed to update Discord reaction count", "error", err)
		}
	})

	// Specify intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsGuildMessageReactions

	// Open the connection
	err = dg.Open()
	if err != nil {
		slog.Error("Error opening Discord connection", "error", err)
		return
	}
}
