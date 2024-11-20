package chain

import (
	"github.com/napuu/gpsp-bot/internal/handlers"
)

type HandlerChain struct {
	rootParser handlers.ContextHandler
}

func NewChainOfResponsibility() *HandlerChain {
	// Initial handler
	telegramParser := &handlers.TelegramTelebotOnTextHandler{}
	discordParser := &handlers.DiscordDiscordGoOnTextHandler{}

	// Basic text message handling
	genericMessageHandler := &handlers.GenericMessageHandler{}

	// URL parsing from the message
	urlParser := &handlers.URLHandler{}

	// Typing indicator for telegram
	telegramTypingHandler := &handlers.TelegramTypingHandler{}
	discordTypingHandler := &handlers.DiscordTypingHandler{}

	// Video processing handlers
	videoCutArgsHandler := &handlers.VideoCutArgsHandler{}
	videoDownloadHandler := &handlers.VideoDownloadHandler{}
	videoPostprocessingHandler := &handlers.VideoPostprocessingHandler{}

	euriborHandler := &handlers.EuriborHandler{}

	// What to do with the results
	markForDeletionHandler := &handlers.MarkForDeletionHandler{}
	markForNaggingHandler := &handlers.MarkForNaggingHandler{}
	constructTextResponseHandler := &handlers.ConstructTextResponseHandler{}

	telegramVideoResponseHandler := &handlers.TelegramVideoResponseHandler{}
	telegramDeleteMessageHandler := &handlers.TelegramDeleteMarkedMessageHandler{}
	telegramTextResponseHandler := &handlers.TelegramTextResponseHandler{}
	telegramTuplillaResponseHandler := &handlers.TelegramTuplillaResponseHandler{}

	discordTextResponseHandler := &handlers.DiscordTextResponseHandler{}

	// Special handler that does not try to call the next handler in the chain
	endOfChainHandler := &handlers.EndOfChainHandler{}

	// Constructing the chain
	telegramParser.SetNext(discordParser)
	discordParser.SetNext(genericMessageHandler)

	genericMessageHandler.SetNext(urlParser)
	urlParser.SetNext(discordTypingHandler)
	discordTypingHandler.SetNext(telegramTypingHandler)

	telegramTypingHandler.SetNext(videoCutArgsHandler)

	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(euriborHandler)

	euriborHandler.SetNext(telegramTuplillaResponseHandler)

	// Response and cleaning handlers
	telegramTuplillaResponseHandler.SetNext(telegramVideoResponseHandler)
	telegramVideoResponseHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(telegramDeleteMessageHandler)

	telegramDeleteMessageHandler.SetNext(telegramTextResponseHandler)
	telegramTextResponseHandler.SetNext(discordTextResponseHandler)
	discordTextResponseHandler.SetNext(endOfChainHandler)

	// Return the initialized chain
	return &HandlerChain{
		rootParser: telegramParser,
	}
}

// Process handles incoming messages through the chain
func (h *HandlerChain) Process(msg *handlers.Context) {
	h.rootParser.Execute(msg)
}
