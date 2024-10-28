package main

import (
	"time"

	"gopkg.in/telebot.v4"
)

type Service int

const (
	Telegram Service = iota + 1
	Discord
	// Matrix // not supported, perhaps at one point
)

type Action int
const (
	Tuplilla Action = iota + 1
	DownloadVideo
	SearchVideo
)

type Context struct {
	service Service
	rawText string
	// Message without action string and
	// possibly related prefixes or suffixes
	parsedText string
	id int
	replyToId int
	isReply bool
	chatId int64
	action Action
	url string

	doneTyping chan struct{}
	gotDubz bool
	dubzNegation chan string
	lastCubeThrownTime time.Time

	telebotContext telebot.Context
	telebot *telebot.Bot

	originalVideoPath string
	possiblyProcessedVideoPath string
	sendVideoSucceeded bool
	startSeconds chan float64
	durationSeconds chan float64
	cutVideoArgsParsed chan bool
	shouldDeleteOriginalMessage bool
	shouldNagAboutOriginalMessage bool

}
const (
	ActionTuplillaString = "tuplilla"
	ActionDownloadVideoString = "dl"
	ActionSearchVideo = "s"
)
