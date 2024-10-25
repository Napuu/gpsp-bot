package main
type HandlerChain struct {
    rootParser handler 
}

func NewHandlerChain() *HandlerChain {

		chainEnd := HandlerLogger(&ChainEnd{})
		telegramResponder := HandlerLogger(&TelegramResponder{})
		telegramResponder.setNext(chainEnd)
    urlParser := HandlerLogger(&URLParser{})
		urlParser.setNext(telegramResponder)
    genericParser := HandlerLogger(&GenericMessageParser{})
    genericParser.setNext(urlParser)
    telegramParser := HandlerLogger(&TelegramMessageParser{})
    telegramParser.setNext(genericParser)

    return &HandlerChain{
        rootParser: telegramParser,
    }
}

// Process handles incoming messages through the chain
func (h *HandlerChain) Process(msg *GenericMessage) {
    h.rootParser.execute(msg)
}
