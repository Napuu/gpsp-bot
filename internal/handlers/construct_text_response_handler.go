package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/napuu/gpsp-bot/internal/version"
)

type ConstructTextResponseHandler struct {
	next ContextHandler
}

func (r *ConstructTextResponseHandler) Execute(m *Context) {
	slog.Debug("Entering ConstructTextResponseHandler")

	var responseText string
	switch m.action {
	case Tuplilla:
		// Telegram is sort of special case of this as
		// it is the only platform with built-in dies
		if m.Service == Telegram {
			if m.gotDubz {
				responseText = "Tuplat tuli 😎"
			} else {
				responseText = "Ei tuplia 😿"
			}
			if m.parsedText != "" {
				if m.gotDubz {
					responseText += ", " + m.parsedText
				} else {
					responseText += ", " + <-m.dubzNegation
				}
			}
			time.Sleep((time.Second * 5) - time.Since(m.lastCubeThrownTime))
		}
	case Ping:
		responseText = "pong"
	case Version:
		responseText = version.Version
	case DownloadVideo:
		if m.shouldNagAboutOriginalMessage {
			responseText = "Hyvä linkki..."
			m.replyToId = m.id
			m.shouldReplyToMessage = true
		}
	case Euribor:
		responseText = fmt.Sprintf(
			`
**Euribor-korot** %s
**12 kk** %.3f %%
**6 kk** %.3f %%
**3 kk** %.3f %%`,
			m.rates.Date.Format("02.01."),
			m.rates.TwelveMonths,
			m.rates.SixMonths,
			m.rates.ThreeMonths,
		)
		if m.Service == Telegram {
			responseText = fmt.Sprintf(
				`
<b>Euribor-korot</b> %s
<b>12 kk</b>: %.3f %%
<b>6 kk</b>: %.3f %%
<b>3 kk</b>: %.3f %%`,
				m.rates.Date.Format("02.01."),
				m.rates.TwelveMonths,
				m.rates.SixMonths,
				m.rates.ThreeMonths,
			)
		}
	}

	m.textResponse = responseText
	r.next.Execute(m)
}

func (u *ConstructTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
