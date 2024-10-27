package main

import "os"

type DeleteDownloadedVideoHandler struct {
	next ContextHandler
}

func (h *DeleteDownloadedVideoHandler) execute(m *Context) {
	if len(m.downloadedVideoPath) > 0 {
		os.Remove(m.downloadedVideoPath)
	}

	h.next.execute(m)
}

func (h *DeleteDownloadedVideoHandler) setNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}