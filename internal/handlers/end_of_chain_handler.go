package handlers

import (
	"log/slog"
)

type EndOfChainHandler struct{}

func (h *EndOfChainHandler) Execute(m *Context) {
	slog.Debug("Entering EndOfChainHandler")
	if m.doneTyping != nil {
		slog.Debug("Closing doneTyping channel")
		close(m.doneTyping)
	}

}

func (h *EndOfChainHandler) SetNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}
