package chain

import (
	"github.com/napuu/gpsp-bot/internal/dayvideo"
	"github.com/napuu/gpsp-bot/internal/handlers"
)

type HandlerChain struct {
	rootParser handlers.ContextHandler
}

func NewChainOfResponsibility(dayVideoScheduler *dayvideo.Scheduler) *HandlerChain {
	onTextHandler := &handlers.OnTextHandler{}
	dayVideoHandler := handlers.NewDayVideoHandler(dayVideoScheduler)

	genericMessageHandler := &handlers.GenericMessageHandler{}

	urlParsingHandler := &handlers.URLParsingHandler{}

	typingHandler := &handlers.TypingHandler{}

	videoCutArgsHandler := &handlers.VideoCutArgsHandler{}
	videoDownloadHandler := &handlers.VideoDownloadHandler{}
	videoPostprocessingHandler := &handlers.VideoPostprocessingHandler{}
	repostDetectionHandler := &handlers.RepostDetectionHandler{}

	euriborHandler := &handlers.EuriborHandler{}

	markForDeletionHandler := &handlers.MarkForDeletionHandler{}
	markForNaggingHandler := &handlers.MarkForNaggingHandler{}
	constructTextResponseHandler := &handlers.ConstructTextResponseHandler{}

	videoResponseHandler := &handlers.VideoResponseHandler{}
	videoStatsHandler := &handlers.VideoStatsHandler{}
	imageResponseHandler := &handlers.ImageResponseHandler{}
	deleteMessageHandler := &handlers.DeleteMessageHandler{}
	textResponseHandler := &handlers.TextResponseHandler{}
	tuplillaResponseHandler := &handlers.TuplillaResponseHandler{}
	hyvaSuomiResponseHandler := &handlers.HyvaSuomiResponseHandler{}
	statsHandler := &handlers.StatsHandler{}

	endOfChainHandler := &handlers.EndOfChainHandler{}

	onTextHandler.SetNext(dayVideoHandler)
	dayVideoHandler.SetNext(genericMessageHandler)

	genericMessageHandler.SetNext(urlParsingHandler)
	urlParsingHandler.SetNext(typingHandler)

	typingHandler.SetNext(videoCutArgsHandler)

	videoCutArgsHandler.SetNext(videoDownloadHandler)
	videoDownloadHandler.SetNext(videoPostprocessingHandler)
	videoPostprocessingHandler.SetNext(repostDetectionHandler)
	repostDetectionHandler.SetNext(euriborHandler)

	euriborHandler.SetNext(statsHandler)
	statsHandler.SetNext(tuplillaResponseHandler)

	tuplillaResponseHandler.SetNext(hyvaSuomiResponseHandler)
	hyvaSuomiResponseHandler.SetNext(videoResponseHandler)
	videoResponseHandler.SetNext(videoStatsHandler)
	videoStatsHandler.SetNext(markForNaggingHandler)
	markForNaggingHandler.SetNext(markForDeletionHandler)
	markForDeletionHandler.SetNext(constructTextResponseHandler)
	constructTextResponseHandler.SetNext(imageResponseHandler)
	imageResponseHandler.SetNext(deleteMessageHandler)

	deleteMessageHandler.SetNext(textResponseHandler)
	textResponseHandler.SetNext(endOfChainHandler)

	return &HandlerChain{
		rootParser: onTextHandler,
	}
}

func (h *HandlerChain) Process(msg *handlers.Context) {
	h.rootParser.Execute(msg)
}
