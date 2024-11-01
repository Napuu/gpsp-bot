package handlers

import (
	"log"
)
type EndOfChainHandler struct {}

func (h *EndOfChainHandler) Execute(m *Context) {
	log.Println("Entering EndOfChainHandler")
	if m.doneTyping != nil {
		log.Println("Closing doneTyping channel")
		close(m.doneTyping)
	}

}

func (h *EndOfChainHandler) SetNext(handler ContextHandler) {
	panic("cannot set next handler on ChainEnd")
}
