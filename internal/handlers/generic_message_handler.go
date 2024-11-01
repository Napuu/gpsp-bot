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
	if textNoPrefix, hasPrefix := strings.CutPrefix(m.rawText, "/"); hasPrefix {
		extractedAction = strings.Split(textNoPrefix, " ")[0]
		textWithoutPrefixOrSuffix = textNoPrefix
	} else if textNoSuffix, hasSuffix := strings.CutSuffix(m.rawText, "!"); hasSuffix {
		split := strings.Split(textNoSuffix, " ")
		extractedAction = split[len(split) - 1]
		textWithoutPrefixOrSuffix = textNoSuffix
	}

	if extractedAction != "" {
		switch extractedAction {
		case ActionDownloadVideoString:
			m.action = DownloadVideo
		case ActionSearchVideo:
			m.action = SearchVideo
		case ActionTuplillaString:
			m.action = Tuplilla
		}

		m.parsedText = strings.Replace(textWithoutPrefixOrSuffix, extractedAction, "", 1)
	}

	mp.next.Execute(m)	
}

func (mp *GenericMessageHandler) SetNext(next ContextHandler) {
	mp.next = next
}