package main

import (
	"github.com/napuu/gpsp-bot/internal/platforms"
)

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	platforms.RunTelegramBot()
}
