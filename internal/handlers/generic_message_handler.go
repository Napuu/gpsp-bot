package handlers

import (
	"strings"
)

// Determines which action user is trying to perform
type GenericMessageHandler struct {
	next ContextHandler
}

func (mp *GenericMessageHandler) Execute(m *Context) {
	var extractedAction string
	var textWithoutPrefixOrSuffix string
	textNoPrefix, hasPrefix := strings.CutPrefix(m.rawText, "/")
	textNoSuffix, hasSuffix := strings.CutSuffix(m.rawText, "!")
	if hasPrefix {
		extractedAction = strings.Split(textNoPrefix, " ")[0]
		textWithoutPrefixOrSuffix = textNoPrefix
	} else if hasSuffix {
		split := strings.Split(textNoSuffix, " ")
		extractedAction = split[len(split) - 1]
		textWithoutPrefixOrSuffix = textNoSuffix
	}

	if (hasPrefix || hasSuffix) && extractedAction != "" {
		switch extractedAction {
		case ActionDownloadVideoString:
			m.action = DownloadVideo
		case ActionSearchVideoString:
			m.action = SearchVideo
		case ActionTuplillaString:
			m.action = Tuplilla
		case ActionPingString:
			m.action = Ping
		}

		m.parsedText = strings.Replace(textWithoutPrefixOrSuffix, extractedAction, "", 1)
	}

	mp.next.Execute(m)	
}

func (mp *GenericMessageHandler) SetNext(next ContextHandler) {
	mp.next = next
}