package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/dayvideo"
	"github.com/napuu/gpsp-bot/internal/doctor"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"github.com/napuu/gpsp-bot/internal/version"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version.GetHumanReadableVersion())
		return
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: gpsp-bot <platform|doctor|day-video> (telegram, discord, doctor, or day-video)")
	}

	command := os.Args[1]
	if command == "doctor" {
		doctor.Run()
		return
	}

	if command == "day-video" {
		outputPath, when, err := parseDayVideoArgs(os.Args[2:])
		if err != nil {
			log.Fatalf("Invalid day-video arguments: %v", err)
		}
		if err := utils.GenerateDayVideo(outputPath, when); err != nil {
			log.Fatalf("Failed to generate day video: %v", err)
		}
		fmt.Println(outputPath)
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

	cfg := config.FromEnv()
	dbPath := filepath.Join(cfg.REPOST_DB_DIR, "repost_fingerprints.duckdb")
	if err := utils.InitRepostDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize stats DB: %v", err)
	}

	var dayVideoScheduler *dayvideo.Scheduler
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch platform {
	case "telegram":
		bot, err := tele.NewBot(tele.Settings{
			Token: token,
		})
		if err != nil {
			log.Fatalf("Failed to initialize Telegram bot: %v", err)
		}

		dayVideoScheduler = dayvideo.NewScheduler(dbPath, bot, nil)
		handlerChain := chain.NewChainOfResponsibility()
		dayVideoScheduler.Start(ctx)

		bot.Handle(tele.OnText, wrapTeleHandler(bot, handlerChain))
		log.Println("Starting Telegram bot...")
		go bot.Start()
		<-ctx.Done()
		bot.Stop()
	case "discord":
		dg, err := discordgo.New("Bot " + token)
		if err != nil {
			log.Fatalf("Failed to initialize Discord session: %v", err)
		}

		dayVideoScheduler = dayvideo.NewScheduler(dbPath, nil, dg)
		handlerChain := chain.NewChainOfResponsibility()
		dayVideoScheduler.Start(ctx)

		dg.AddHandler(wrapDiscoHandler(handlerChain))
		if err := dg.Open(); err != nil {
			log.Fatalf("Failed to start Discord bot: %v", err)
		}
		defer dg.Close()
		log.Println("Starting Discord bot...")
		<-ctx.Done()
	}
}

// wrapTeleHandler wraps the chain for Telegram.
func wrapTeleHandler(bot *tele.Bot, handlerChain *chain.HandlerChain) func(c tele.Context) error {
	return func(c tele.Context) error {
		handlerChain.Process(&handlers.Context{TelebotContext: c, Telebot: bot, Service: handlers.Telegram})
		return nil
	}
}

// wrapDiscoHandler wraps the chain for Discord.
func wrapDiscoHandler(handlerChain *chain.HandlerChain) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		handlerChain.Process(&handlers.Context{DiscordSession: s, DiscordMessage: m, Service: handlers.Discord})
	}
}

func parseDayVideoArgs(args []string) (outputPath string, when time.Time, err error) {
	outputPath = filepath.Join(utils.DayVideoTmpDir(), "output.mp4")
	when = time.Now().UTC()

	for _, arg := range args {
		parsed, parseErr := time.ParseInLocation("2006-01-02", arg, time.UTC)
		if parseErr == nil {
			when = parsed
			continue
		}
		outputPath = arg
	}

	return outputPath, when, nil
}
