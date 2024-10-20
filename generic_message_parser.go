package main

import (
	"strings"
)

// Determines which action user is trying to perform
type GenericMessageParser struct {
	next handler
}

func (mp *GenericMessageParser) execute(m *GenericMessage) {
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

	mp.next.execute(m)	
}

func (mp *GenericMessageParser) setNext(next handler) {
	mp.next = next
}