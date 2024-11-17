package handlers

import (
	"log/slog"
	"regexp"
)

type URLHandler struct {
	next ContextHandler
}

func (u *URLHandler) Execute(m *Context) {
	slog.Debug("Entering URLHandler")
	urlRegex := `https?://[a-zA-Z0-9.-]+(:[0-9]{1,5})?(/[a-zA-Z0-9./?=&_@+!*(),;%~-]*)?`
	re := regexp.MustCompile(urlRegex)
	match := re.FindString(m.parsedText)

	m.url = match

	u.next.Execute(m)
}

func (u *URLHandler) SetNext(next ContextHandler) {
	u.next = next
}
