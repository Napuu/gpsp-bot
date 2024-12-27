package chain

import (
	"github.com/napuu/gpsp-bot/internal/handlers"
)

type HandlerChain struct {
	rootParser handlers.ContextHandler
}

func NewChainOfResponsibility() *HandlerChain {
	onTextHandler := &handlers.OnTextHandler{}

	genericMessageHandler := &handlers.GenericMessageHandler{}

	urlParsingHandler := &handlers.URLParsingHandler{}

	typingHandler := &handlers.TypingHandler{}

	videoCutArgsHandler := &handlers.VideoCutArgsHandler{}
	videoDownloadHandler := &handlers.VideoDownloadHandler{}
	videoPostprocessingHandler := &handlers.VideoPostprocessingHandler{}

	euriborHandler := &handlers.EuriborHandler{}

	markForDeletionHandler := &handlers.MarkForDeletionHandler{}
	markForNaggingHandler := &handlers.MarkForNaggingHandler{}
	constructTextResponseHandler := &handlers.ConstructTextResponseHandler{}

	videoResponseHandler := &handlers.VideoResponseHandler{}
	deleteMessageHandler := &handlers.DeleteMessageHandler{}
	textResponseHandler := &handlers.TextResponseHandler{}
	tuplillaResponseHandler := &handlers.TuplillaResponseHandler{}

	endOfChainHandler := &handlers.EndOfChainHandler{}

	onTextHandler.SetNext(genericMessageHandler)

	genericMessageHandler.SetNext(urlParsingHandler)
	urlParsingHandler.SetNext(typingHandler)

	typingHandler.SetNext(videoCutArgsHandler)

	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(euriborHandler)

	euriborHandler.SetNext(tuplillaResponseHandler)

	tuplillaResponseHandler.SetNext(videoResponseHandler)
	videoResponseHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(deleteMessageHandler)

	deleteMessageHandler.SetNext(textResponseHandler)
	textResponseHandler.SetNext(endOfChainHandler)

	return &HandlerChain{
		rootParser: onTextHandler,
	}
}

func (h *HandlerChain) Process(msg *handlers.Context) {
	h.rootParser.Execute(msg)
}
