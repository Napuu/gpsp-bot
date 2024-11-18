package handlers

import (
	"time"

	"github.com/napuu/gpsp-bot/pkg/utils"
	"gopkg.in/telebot.v4"
)

type Service int

const (
	Telegram Service = iota + 1
	Discord
	// Matrix // not supported, perhaps at one point
)

type ContextHandler interface {
	Execute(*Context)
	SetNext(ContextHandler)
}

type Action string

const (
	Tuplilla      Action = "tuplilla"
	DownloadVideo Action = "dl"
	SearchVideo   Action = "s"
	Ping          Action = "ping"
	Euribor       Action = "euribor"
)

const (
	TuplillaDescription      ActionDescription = "Tuplilla..."
	DownloadVideoDescription ActionDescription = "Lataa video"
	SearchVideoDescription   ActionDescription = "Etsi ja lataa video YouTubesta"
	PingDescription          ActionDescription = "Ping..."
	EuriborDescription       ActionDescription = "Tuoreet Euribor-korot"
)

func VisibleCommands() map[Action]ActionDescription {
	return map[Action]ActionDescription{
		Tuplilla:      TuplillaDescription,
		DownloadVideo: DownloadVideoDescription,
		Euribor:       EuriborDescription,
	}
}

type Context struct {
	Service Service
	rawText string
	// Message without action string and
	// possibly related prefixes or suffixes
	parsedText string
	id         int64
	replyToId  int64
	// Must store separate from replyToId as
	// replyToId = 0 might refer to first message
	// or to no message at all
	shouldReplyToMessage bool
	isReply              bool
	chatId               int64
	action               Action
	url                  string

	doneTyping         chan struct{}
	gotDubz            bool
	dubzNegation       chan string
	lastCubeThrownTime time.Time

	rates utils.LatestEuriborRates

	TelebotContext telebot.Context
	Telebot        *telebot.Bot

	originalVideoPath             string
	possiblyProcessedVideoPath    string
	textResponse                  string
	sendVideoSucceeded            bool
	startSeconds                  chan float64
	durationSeconds               chan float64
	cutVideoArgsParsed            chan bool
	shouldDeleteOriginalMessage   bool
	shouldNagAboutOriginalMessage bool
}
