package main

import "log"

type VideoDownloadHandler struct {
	next ContextHandler
}

func (u *VideoDownloadHandler) execute(m *Context) {
	log.Println("Entering VideoDownloadHandler")
	if len(m.url) > 0 && m.action == DownloadVideo {
		path := downloadVideo(m.url, 5)

		m.downloadedVideoPath = path
	}
	u.next.execute(m)
}

func (u *VideoDownloadHandler) setNext(next ContextHandler) {
	u.next = next
}