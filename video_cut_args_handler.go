package main

import (
	"log"
	"strings"
)

type VideoCutArgsHandler struct {
	next ContextHandler
}

func (u *VideoCutArgsHandler) execute(m *Context) {
	log.Println("Entering VideoCutArgsHandler")
	leftover := strings.Replace(m.parsedText, m.url, "", 1)

	m.startSeconds = make(chan float64)
	m.durationSeconds = make(chan float64)

	MIN_LEFTOVER_LEN_TO_CONSIDER := 2
	m.cutVideoArgsParsed = make(chan bool)
	go func() {
		if m.action == DownloadVideo && len(leftover) > MIN_LEFTOVER_LEN_TO_CONSIDER {
			startSeconds, durationSeconds, err := parseCutArgs(leftover)
			if err != nil {
				log.Println("Failed", err)
				m.cutVideoArgsParsed <- false
			} else {
				log.Println("Got args back", startSeconds, durationSeconds)
				m.cutVideoArgsParsed <- true
				m.startSeconds <- startSeconds
				m.durationSeconds <- durationSeconds
			}
		} else {
			m.cutVideoArgsParsed <- false
		}
	}()

	u.next.execute(m)
}

func (u *VideoCutArgsHandler) setNext(next ContextHandler) {
	u.next = next
}