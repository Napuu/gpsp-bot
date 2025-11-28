package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/napuu/gpsp-bot/internal/platforms"
	"github.com/napuu/gpsp-bot/internal/version"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version.Version)
		return
	}

	platforms.EnsureBotCanStart()
	platforms.VerifyEnabledCommands()
	if len(os.Args) == 1 {
		log.Fatalf("Usage: gpsp-bot <telegram/discord>")
	}

	sc := make(chan os.Signal, 1)
	switch os.Args[1] {
	case "telegram":
		slog.Info("Starting Telegram bot...")
		platforms.RunTelegramBot()
		slog.Info("Telegram bot started!")
	case "discord":
		slog.Info("Starting Discord bot...")
		platforms.RunDiscordBot()
		slog.Info("Discord bot started!")
	}
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
