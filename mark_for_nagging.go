package main

import "log"

type MarkForNaggingHandler struct {
	next ContextHandler
}

func (u *MarkForNaggingHandler) execute(m *Context) {
	log.Println("Entering MarkForNaggingHandler")
	if m.action == DownloadVideo && !m.sendVideoSucceeded {
		log.Println("shouldNag set true")
		m.shouldNagAboutOriginalMessage = true
	}

	u.next.execute(m)
}

func (u *MarkForNaggingHandler) setNext(next ContextHandler) {
	u.next = next
}