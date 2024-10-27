package main
type HandlerChain struct {
    rootParser ContextHandler 
}

func NewChainOfResponsibility() *HandlerChain {

    telegramParser := &TelegramIncomingContextHandler{}

    genericMessageHandler := &GenericMessageHandler{}
    urlParser := &URLHandler{}

    telegramTypingHandler := &TelegramTypingHandler{}

    videoDownloadHandler := &VideoDownloadHandler{}
    markForDeletionHandler := &MarkForDeletionHandler{}
    markForNaggingHandler := &MarkForNaggingHandler{}

    telegramVideoResponseHandler := &TelegramVideoResponseHandler{}
    telegramDeleteMessageHandler := &TelegramDeleteMarkedMessageHandler{}
    telegramTextResponseHandler := &TelegramTextResponseHandler{}
    telegramTuplillaResponseHandler := &TelegramTuplillaResponseHandler{}

    endOfChainHandler := &EndOfChainHandler{}

    // Constructing the chain
    telegramParser.setNext(genericMessageHandler)

    genericMessageHandler.setNext(urlParser)
    urlParser.setNext(telegramTypingHandler)

    telegramTypingHandler.setNext(videoDownloadHandler)

    videoDownloadHandler.setNext(markForDeletionHandler)
    markForDeletionHandler.setNext(markForNaggingHandler)
    markForNaggingHandler.setNext(telegramTuplillaResponseHandler)

    telegramTuplillaResponseHandler.setNext(telegramVideoResponseHandler)
    telegramVideoResponseHandler.setNext(telegramDeleteMessageHandler)
    telegramDeleteMessageHandler.setNext(telegramTextResponseHandler)
    telegramTextResponseHandler.setNext(endOfChainHandler)

    return &HandlerChain{
        rootParser: telegramParser,
    }
}

// Process handles incoming messages through the chain
func (h *HandlerChain) Process(msg *Context) {
    h.rootParser.execute(msg)
}
