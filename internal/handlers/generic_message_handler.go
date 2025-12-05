package handlers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/napuu/gpsp-bot/internal/config"
)

type GenericMessageHandler struct {
	next ContextHandler
}

func (mp *GenericMessageHandler) Execute(m *Context) {
	slog.Debug("rawText: " + m.rawText)
	var extractedAction string
	var textWithoutPrefixOrSuffix string

	prefixes := []string{"/", "!"}
	textNoPrefix := ""
	hasPrefix := false
	textNoSuffix, hasSuffix := strings.CutSuffix(m.rawText, "!")

	for _, prefix := range prefixes {
		if strings.HasPrefix(m.rawText, prefix) {
			textNoPrefix, hasPrefix = strings.CutPrefix(m.rawText, prefix)
			break
		}
	}

	if hasPrefix {
		extractedAction = strings.Split(textNoPrefix, " ")[0]
		textWithoutPrefixOrSuffix = textNoPrefix
	} else if hasSuffix {
		split := strings.Split(textNoSuffix, " ")
		extractedAction = split[len(split)-1]
		textWithoutPrefixOrSuffix = textNoSuffix
	}

	if (hasPrefix || hasSuffix) && extractedAction != "" && strings.Contains(config.FromEnv().ENABLED_FEATURES, extractedAction) {
		switch Action(extractedAction) {
		case DownloadVideo:
			m.action = DownloadVideo
		case Tuplilla:
			m.action = Tuplilla
		case Ping:
			m.action = Ping
		case Euribor:
			m.action = Euribor
		case Version:
			m.action = Version
		}

		m.parsedText = strings.TrimSpace(strings.Replace(textWithoutPrefixOrSuffix, extractedAction, "", 1))
	}

	if m.action != "" {
		slog.Info(fmt.Sprintf("Command '%s' received", m.action))
	}

	mp.next.Execute(m)
}

func (mp *GenericMessageHandler) SetNext(next ContextHandler) {
	mp.next = next
}
