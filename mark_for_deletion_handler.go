package main

import "log"

type MarkForDeletionHandler struct {
	next ContextHandler
}

func (u *MarkForDeletionHandler) execute(m *Context) {
	log.Println("Entering MarkForDeletionHandler")
	if m.action == DownloadVideo && m.sendVideoSucceeded {
		m.shouldDeleteOriginalMessage = true
	}

	u.next.execute(m)
}

func (u *MarkForDeletionHandler) setNext(next ContextHandler) {
	u.next = next
}