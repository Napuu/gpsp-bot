package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tele "gopkg.in/telebot.v4"
	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/doctor"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"github.com/napuu/gpsp-bot/internal/version"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version.GetHumanReadableVersion())
		return
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: gpsp-bot <platform|doctor> (telegram, discord, or doctor)")
	}

	command := os.Args[1]
	if command == "doctor" {
		doctor.Run()
		return
	}

	platform := command
	if platform != "telegram" && platform != "discord" {
		log.Fatal("Platform must be either 'telegram' or 'discord'")
	}

	enabledFeatures := config.EnabledFeatures()
	if len(enabledFeatures) == 0 || (len(enabledFeatures) == 1 && enabledFeatures[0] == "") {
		log.Fatal("ENABLED_FEATURES environment variable is required")
	}

	var token string
	switch platform {
	case "telegram":
		token = os.Getenv("TELEGRAM_TOKEN")
		if token == "" {
			log.Fatal("TELEGRAM_TOKEN environment variable is required")
		}
	case "discord":
		token = os.Getenv("DISCORD_TOKEN")
		if token == "" {
			log.Fatal("DISCORD_TOKEN environment variable is required")
		}
	}

	// Initialize the chain of responsibility
	chain := chain.NewChainOfResponsibility()

	// Initialize the platform
	dbPath := filepath.Join("/tmp/repost-db", "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize stats DB: %v", err)
	}

	switch platform {
	case "telegram":
		bot, err := tele.NewBot(tele.Settings{
			Token: token,
		})
		if err != nil {
			log.Fatalf("Failed to initialize Telegram bot: %v", err)
		}

		bot.Handle(tele.OnText, wrapTeleHandler(bot, chain))
		log.Println("Starting Telegram bot...")
		bot.Start()
	case "discord":
		dg, err := discordgo.New("Bot " + token)
		if err != nil {
			log.Fatalf("Failed to initialize Discord session: %v", err)
		}

		dg.AddHandler(wrapDiscoHandler(chain))
		if err := dg.Open(); err != nil {
			log.Fatalf("Failed to start Discord bot: %v", err)
		}
		defer dg.Close()
		log.Println("Starting Discord bot...")
		<-make(chan struct{})
	}
}

// wrapTeleHandler wraps the chain for Telegram.
func wrapTeleHandler(bot *tele.Bot, chain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		chain.Process(&handlers.Context{TelebotContext: c, Telebot: bot, Service: handlers.Telegram})
		return nil
	}
}

// wrapDiscoHandler wraps the chain for Discord.
func wrapDiscoHandler(chain *chain.HandlerChain) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		chain.Process(&handlers.Context{DiscordSession: s, DiscordMessage: m, Service: handlers.Discord})
	}
}