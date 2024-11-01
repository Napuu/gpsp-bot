package main
type HandlerChain struct {
    rootParser ContextHandler 
}

func NewChainOfResponsibility() *HandlerChain {

    telegramParser := &TelegramTelebotOnTextHandler{}

    genericMessageHandler := &GenericMessageHandler{}
    urlParser := &URLHandler{}

    telegramTypingHandler := &TelegramTypingHandler{}

    videoDownloadHandler := &VideoDownloadHandler{}
    videoPostprocessingHandler := &VideoPostprocessingHandler{}
    videoCutArgsHandler := &VideoCutArgsHandler{}
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

    telegramTypingHandler.setNext(videoCutArgsHandler)

    videoCutArgsHandler.setNext(videoDownloadHandler)
    videoDownloadHandler.setNext(videoPostprocessingHandler)
    videoPostprocessingHandler.setNext(telegramTuplillaResponseHandler)

    telegramTuplillaResponseHandler.setNext(telegramVideoResponseHandler)
    telegramVideoResponseHandler.setNext(markForNaggingHandler)
    markForNaggingHandler.setNext(markForDeletionHandler)
    markForDeletionHandler.setNext(telegramDeleteMessageHandler)
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
