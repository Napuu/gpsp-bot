package main

import (
	"fmt"
	"log"
	"time"
)

type TelegramTextResponseHandler struct {
	next ContextHandler
}

func (r *TelegramTextResponseHandler) execute(m *Context) {
	log.Println("Entering TelegramTextResponseHandler")
	if m.service == Telegram {
		if m.shouldNagAboutOriginalMessage {
			err := m.telebotContext.Reply("Hyvä linkki...")
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
			m.telebotContext.Send(dubzResultMessage)
		}

	}

	r.next.execute(m)
}

func (u *TelegramTextResponseHandler) setNext(next ContextHandler) {
	u.next = next
}