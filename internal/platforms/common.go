package platforms

import (
	"fmt"
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

func actionExists(action string) bool {
	_, exists := handlers.ActionMap[handlers.Action(action)]
	return exists
}

func VerifyEnabledCommands() {
	for _, action := range config.EnabledFeatures() {
		if actionExists(action) {
			slog.Info(fmt.Sprintf("Enabled action %s: '%s'!", action, handlers.ActionMap[handlers.Action(action)]))
		} else {
			panic(fmt.Sprintf("Action '%s' does not exist", action))
		}

	}
}

func EnsureBotCanStart() {
	utils.EnsureTmpDirExists(config.FromEnv().YTDLP_TMP_DIR)
	utils.EnsureTmpDirExists(config.FromEnv().EURIBOR_GRAPH_DIR)
	utils.EnsureTmpDirExists(config.FromEnv().EURIBOR_CSV_DIR)
}
