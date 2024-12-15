package handlers

import (
	"log/slog"
	"time"
)

type TypingHandler struct {
	next ContextHandler
}

func (t *TypingHandler) Execute(m *Context) {
	slog.Debug("Entering TypingHandler")
	if m.action != "" {
		m.doneTyping = make(chan struct{})

		go func() {
			ticker := time.NewTicker(4 * time.Second)
			defer ticker.Stop()

			m.SendTyping()

			for {
				select {
				case <-m.doneTyping:
					return
				case <-ticker.C:
					m.SendTyping()
				}
			}
		}()
	}

	t.next.Execute(m)
}

func (t *TypingHandler) SetNext(next ContextHandler) {
	t.next = next
}
