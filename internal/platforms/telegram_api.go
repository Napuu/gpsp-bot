package platforms

import (
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"

	tele "gopkg.in/telebot.v4"
)

func TelebotCompatibleVisibleCommands() []tele.Command {
	commands := make([]tele.Command, 0, len(config.EnabledFeatures()))
	for _, action := range config.EnabledFeatures() {
		if action == "" || config.IsBackgroundFeature(action) {
			continue
		}
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
