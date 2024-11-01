package handlers

import (
	"fmt"
	"log"
	"time"
)

type TelegramTextResponseHandler struct {
	next ContextHandler
}

func (r *TelegramTextResponseHandler) Execute(m *Context) {
	log.Println("Entering TelegramTextResponseHandler")
	if m.Service == Telegram {
		if m.shouldNagAboutOriginalMessage {
			err := m.TelebotContext.Reply("Hyvä linkki...")
			if err != nil {
				log.Println(err)
			}
		}
		
		if m.action == Tuplilla {
			var dubzResultMessage string
			if m.gotDubz {
				dubzResultMessage = fmt.Sprintf("Tuplat tuli 😎, %s", m.parsedText)
			} else {
				negated := <- m.dubzNegation
				dubzResultMessage = fmt.Sprintf("Ei tuplia 😿, %s", negated)
			}
			time.Sleep((time.Second * 5) - time.Since(m.lastCubeThrownTime))
			m.TelebotContext.Send(dubzResultMessage)
		}

	}

	r.next.Execute(m)
}

func (u *TelegramTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}