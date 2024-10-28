package main

import (
	"fmt"
	"log"
)

type VideoDownloadHandler struct {
	next ContextHandler
}

func (u *VideoDownloadHandler) execute(m *Context) {
	log.Println("Entering VideoDownloadHandler", m.action)
	if m.action == DownloadVideo || m.action == SearchVideo {
		var videoString = m.url
		if m.action == SearchVideo {
			videoString = fmt.Sprintf("ytsearch:\"%s\"", videoString)
		}
		path := downloadVideo(videoString, 5)

		m.originalVideoPath = path
	}
	u.next.execute(m)
}

func (u *VideoDownloadHandler) setNext(next ContextHandler) {
	u.next = next
}