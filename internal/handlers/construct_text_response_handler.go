package handlers

import (
	"fmt"
	"log"
	"time"
)

type ConstructTextResponseHandler struct {
	next ContextHandler
}

func (r *ConstructTextResponseHandler) Execute(m *Context) {
	log.Println("Entering ConstructTextResponseHandler")

	var responseText string
	if m.action == Tuplilla {
		if m.gotDubz {
			responseText = fmt.Sprintf("Tuplat tuli 😎, %s", m.parsedText)
		} else {
			negated := <-m.dubzNegation
			responseText = fmt.Sprintf("Ei tuplia 😿, %s", negated)
		}
		time.Sleep((time.Second * 5) - time.Since(m.lastCubeThrownTime))
	}

	if m.action == Ping {
		responseText = "pong"
	}

	if (m.action == DownloadVideo || m.action == SearchVideo) && m.shouldNagAboutOriginalMessage {
		responseText = "Hyvä linkki..."
		m.replyToId = m.id
		m.shouldReplyToMessage = true
	}

	m.textResponse = responseText
	r.next.Execute(m)
}

func (u *ConstructTextResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
