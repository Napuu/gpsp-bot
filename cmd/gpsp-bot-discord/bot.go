package main

import (
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/platforms"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	platforms.RunDiscordBot()
}
