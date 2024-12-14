package chain

import (
	"github.com/napuu/gpsp-bot/internal/handlers"
)

type HandlerChain struct {
	rootParser handlers.ContextHandler
}

func NewChainOfResponsibility() *HandlerChain {
	// Initial handler
	onTextHandler := &handlers.OnTextHandler{}

	// Basic text message handling
	genericMessageHandler := &handlers.GenericMessageHandler{}

	// URL parsing from the message
	urlParsingHandler := &handlers.URLParsingHandler{}

	// Typing indicator for telegram
	typingHandler := &handlers.TypingHandler{}

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
	deleteMessageHandler := &handlers.DeleteMessageHandler{}
	textResponseHandler := &handlers.TextResponseHandler{}
	tuplillaResponseHandler := &handlers.TuplillaResponseHandler{}

	// Special handler that does not try to call the next handler in the chain
	endOfChainHandler := &handlers.EndOfChainHandler{}

	// Constructing the chain
	onTextHandler.SetNext(genericMessageHandler)

	genericMessageHandler.SetNext(urlParsingHandler)
	urlParsingHandler.SetNext(typingHandler)

	typingHandler.SetNext(videoCutArgsHandler)

	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(euriborHandler)

	euriborHandler.SetNext(tuplillaResponseHandler)

	// Response and cleaning handlers
	tuplillaResponseHandler.SetNext(telegramVideoResponseHandler)
	telegramVideoResponseHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(deleteMessageHandler)

	deleteMessageHandler.SetNext(textResponseHandler)
	textResponseHandler.SetNext(endOfChainHandler)

	// Return the initialized chain
	return &HandlerChain{
		rootParser: onTextHandler,
	}
}

// Process handles incoming messages through the chain
func (h *HandlerChain) Process(msg *handlers.Context) {
	h.rootParser.Execute(msg)
}
