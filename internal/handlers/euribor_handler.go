package handlers

import (
	"log/slog"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/repository"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type EuriborHandler struct {
	next ContextHandler
}

func (t *EuriborHandler) Execute(m *Context) {
	slog.Debug("Entering EuriborHandler")

	if m.action == Euribor {
		db, _ := repository.InitializeDB()
		cachedRates, _ := repository.GetCachedRates(db)

		var data utils.EuriborData

		if cachedRates != nil {
			slog.Debug("Using cached Euribor rates")
			data.Latest = cachedRates.Value
			// Still load CSV and history for charting
			tempData := utils.GetEuriborData()
			data = tempData
		} else {
			slog.Debug("Fetching fresh Euribor rates")
			data = utils.GetEuriborData()

			repository.InsertRates(db, repository.RateCache{
				Value:       data.Latest,
				LastFetched: time.Now(),
			})
		}
		tmpPath := config.FromEnv().EURIBOR_GRAPH_DIR
		var path = tmpPath + "/" + time.Now().Format("2006-01-02") + ".jpg"
		utils.GenerateLine(data.History, path)

		m.rates = data.Latest
		m.finalImagePath = path
	}

	t.next.Execute(m)
}

func (t *EuriborHandler) SetNext(next ContextHandler) {
	t.next = next
}
