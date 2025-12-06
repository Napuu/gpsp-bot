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

// LogLevel is set at build time via ldflags (debug, info, warn, error)
// e.g. -ldflags="-X main.LogLevel=debug"
var LogLevel = "info"

func main() {
	// Configure log level
	var level slog.Level
	switch LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))

	if len(os.Args) >= 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version.GetHumanReadableVersion())
		return
	}

	platforms.EnsureBotCanStart()
	platforms.VerifyEnabledCommands()
	if len(os.Args) == 1 {
		log.Fatalf("Usage: gpsp-bot <telegram/discord/matrix>")
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
	case "matrix":
		slog.Info("Starting Matrix bot...")
		platforms.RunMatrixBot()
		slog.Info("Matrix bot started!")
	}
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
