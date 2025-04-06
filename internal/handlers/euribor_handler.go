package handlers

import (
	"log/slog"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/repository"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
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
		chatId := tele.ChatID(utils.S2I(m.chatId))
		m.Telebot.Send(chatId, &tele.Photo{File: tele.FromDisk(path)})

		m.rates = data.Latest
		m.chartPath = path
	}

	t.next.Execute(m)
}

func (t *EuriborHandler) SetNext(next ContextHandler) {
	t.next = next
}
