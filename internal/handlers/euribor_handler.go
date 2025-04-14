package handlers

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type EuriborHandler struct {
	next ContextHandler
}

func (t *EuriborHandler) Execute(m *Context) {
	slog.Debug("Entering EuriborHandler")

	if m.action == Euribor {
		tmpPath := config.FromEnv().EURIBOR_CSV_DIR
		euriborExportFile := tmpPath + "/" + uuid.New().String() + ".csv"
		var path = config.FromEnv().EURIBOR_GRAPH_DIR + "/" + uuid.New().String() + ".jpg"

		if utils.ShouldFetchCSV(tmpPath, 15*time.Minute) {
			utils.DownloadEuriborCSVFile(euriborExportFile)
		}
		data := utils.GetRatesFromCSV(tmpPath, time.Now().AddDate(0, -1, 0))

		utils.GenerateLine(data, path)

		m.finalImagePath = path
		m.rates = data[0]
	}

	t.next.Execute(m)
}

func (t *EuriborHandler) SetNext(next ContextHandler) {
	t.next = next
}
