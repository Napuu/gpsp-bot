package chain

import (
	"github.com/napuu/gpsp-bot/internal/handlers"
)

type HandlerChain struct {
    rootParser handlers.ContextHandler 
}


func NewChainOfResponsibility() *HandlerChain {

    telegramParser := &handlers.TelegramTelebotOnTextHandler{}

    genericMessageHandler := &handlers.GenericMessageHandler{}
    urlParser := &handlers.URLHandler{}

    telegramTypingHandler := &handlers.TelegramTypingHandler{}

    videoDownloadHandler := &handlers.VideoDownloadHandler{}
    videoPostprocessingHandler := &handlers.VideoPostprocessingHandler{}
    videoCutArgsHandler := &handlers.VideoCutArgsHandler{}
    markForDeletionHandler := &handlers.MarkForDeletionHandler{}
    markForNaggingHandler := &handlers.MarkForNaggingHandler{}

    telegramVideoResponseHandler := &handlers.TelegramVideoResponseHandler{}
    telegramDeleteMessageHandler := &handlers.TelegramDeleteMarkedMessageHandler{}
    telegramTextResponseHandler := &handlers.TelegramTextResponseHandler{}
    telegramTuplillaResponseHandler := &handlers.TelegramTuplillaResponseHandler{}

    endOfChainHandler := &handlers.EndOfChainHandler{}

    // Constructing the chain
    telegramParser.SetNext(genericMessageHandler)

    genericMessageHandler.SetNext(urlParser)
    urlParser.SetNext(telegramTypingHandler)

    telegramTypingHandler.SetNext(videoCutArgsHandler)

    videoCutArgsHandler.SetNext(videoDownloadHandler)
    videoDownloadHandler.SetNext(videoPostprocessingHandler)
    videoPostprocessingHandler.SetNext(telegramTuplillaResponseHandler)

    telegramTuplillaResponseHandler.SetNext(telegramVideoResponseHandler)
    telegramVideoResponseHandler.SetNext(markForNaggingHandler)
    markForNaggingHandler.SetNext(markForDeletionHandler)
    markForDeletionHandler.SetNext(telegramDeleteMessageHandler)
    telegramDeleteMessageHandler.SetNext(telegramTextResponseHandler)
    telegramTextResponseHandler.SetNext(endOfChainHandler)

    return &HandlerChain{
        rootParser: telegramParser,
    }
}

// Process handles incoming messages through the chain
func (h *HandlerChain) Process(msg *handlers.Context) {
    h.rootParser.Execute(msg)
}
