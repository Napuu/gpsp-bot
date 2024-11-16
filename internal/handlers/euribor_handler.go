package handlers

import (
	"log/slog"
	"time"

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
		if cachedRates != nil {
			m.rates = cachedRates.Value
		} else {
			newRates := utils.GetEuriborRates()

			repository.InsertRates(db, repository.RateCache{Value: newRates, LastFetched: time.Now()})

			m.rates = newRates
		}
	}

	t.next.Execute(m)
}

func (t *EuriborHandler) SetNext(next ContextHandler) {
	t.next = next
}
