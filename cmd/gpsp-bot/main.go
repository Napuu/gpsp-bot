package main

import (
	"log/slog"
	"os"

	"github.com/napuu/gpsp-bot/internal/platforms"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	platforms.EnsureBotCanStart()
	switch os.Args[1] {
	case "telegram":
		platforms.RunTelegramBot()
	case "discord":
		platforms.RunDiscordBot()
	case "matrix":
		platforms.RunMatrixBot()
	}
}
