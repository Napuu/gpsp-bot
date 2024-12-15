package main

import (
	"github.com/napuu/gpsp-bot/internal/platforms"
)

func main() {
	platforms.EnsureBotCanStart()
	platforms.RunDiscordBot()
}
