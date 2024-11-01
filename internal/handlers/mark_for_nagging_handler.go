package handlers

import "log"

type MarkForNaggingHandler struct {
	next ContextHandler
}

func (u *MarkForNaggingHandler) Execute(m *Context) {
	log.Println("Entering MarkForNaggingHandler")
	if m.action == DownloadVideo && !m.sendVideoSucceeded {
		log.Println("shouldNag set true")
		m.shouldNagAboutOriginalMessage = true
	}

	u.next.Execute(m)
}

func (u *MarkForNaggingHandler) SetNext(next ContextHandler) {
	u.next = next
}
