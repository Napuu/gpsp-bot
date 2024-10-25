package main

import (
	"fmt"
	"regexp"
)

type URLParser struct {
	next handler
}

func (u *URLParser) execute(m *GenericMessage) {
	urlRegex := `https?://[a-zA-Z0-9./?=&_-]+`
	re := regexp.MustCompile(urlRegex)
	match := re.FindString(m.parsedText)

	m.url = match

	if len(m.url) > 0 {
		fmt.Println("attempting download for fun")
		downloadVideo(m.url, 20)
	}

	fmt.Println(m.url)

	u.next.execute(m)
}

func (u *URLParser) setNext(next handler) {
	u.next = next
}
