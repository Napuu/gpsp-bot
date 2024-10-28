package main

import (
	"log"
	"regexp"
)

type URLHandler struct {
	next ContextHandler
}

func (u *URLHandler) execute(m *Context) {
	log.Println("Entering URLHandler")
	urlRegex := `https?://[a-zA-Z0-9.-]+(:[0-9]{1,5})?(/[a-zA-Z0-9./?=&_-]*)?`
	re := regexp.MustCompile(urlRegex)
	match := re.FindString(m.parsedText)

	m.url = match

	u.next.execute(m)
}

func (u *URLHandler) setNext(next ContextHandler) {
	u.next = next
}
