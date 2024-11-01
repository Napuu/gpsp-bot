package handlers

import "os"

type DeleteDownloadedVideoHandler struct {
	next ContextHandler
}

func (h *DeleteDownloadedVideoHandler) Execute(m *Context) {
	if len(m.originalVideoPath) > 0 {
		os.Remove(m.originalVideoPath)
	}
	if len(m.possiblyProcessedVideoPath ) > 0 {
		os.Remove(m.possiblyProcessedVideoPath)
	}

	h.next.Execute(m)
}

func (h *DeleteDownloadedVideoHandler) SetNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}