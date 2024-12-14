package handlers

import (
	"time"

	dg "github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
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
type ActionDescription string

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
	// The original message without any parsing
	// (except on Telegram events, the possible "@<botname>"" is removed)
	rawText string
	// Message without action string and
	// possibly related prefixes or suffixes
	parsedText string
	id         string // Some services use string, some int, some int64. They're now strings at our context.
	replyToId  string
	// Must store separate from replyToId as
	// replyToId = 0 might refer to first message
	// or to no message at all
	shouldReplyToMessage bool
	isReply              bool
	chatId               string
	action               Action
	url                  string

	doneTyping         chan struct{}
	gotDubz            bool
	dubzNegation       chan string
	lastCubeThrownTime time.Time

	rates utils.LatestEuriborRates

	TelebotContext tele.Context
	Telebot        *tele.Bot

	DiscordSession *dg.Session
	DiscordMessage *dg.MessageCreate

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
