package platforms

import (
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

func EnsureBotCanStart() {
	utils.EnsureTmpDirExists(config.FromEnv().YTDLP_TMP_DIR)
}
