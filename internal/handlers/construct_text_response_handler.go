package handlers

import (
	"fmt"
	"log/slog"
	"time"
)

type ConstructTextResponseHandler struct {
	next ContextHandler
}

func (r *ConstructTextResponseHandler) Execute(m *Context) {
	slog.Debug("Entering ConstructTextResponseHandler")

	var responseText string
	switch m.action {
	case Tuplilla:
		if m.gotDubz {
			responseText = fmt.Sprintf("Tuplat tuli 😎, %s", m.parsedText)
		} else {
			negated := <-m.dubzNegation
			responseText = fmt.Sprintf("Ei tuplia 😿, %s", negated)
		}
		time.Sleep((time.Second * 5) - time.Since(m.lastCubeThrownTime))
	case Ping:
		responseText = "pong"
	case DownloadVideo:
		fallthrough
	case SearchVideo:
		if m.shouldNagAboutOriginalMessage {
			responseText = "Hyvä linkki..."
			m.replyToId = m.id
			m.shouldReplyToMessage = true
		}
	case Euribor:
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

	m.textResponse = responseText
	r.next.Execute(m)
}

func (u *ConstructTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
