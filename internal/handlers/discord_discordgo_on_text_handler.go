package handlers

import "log/slog"

type DiscordDiscordGoOnTextHandler struct {
	next ContextHandler
}

func (discordMessageParser *DiscordDiscordGoOnTextHandler) Execute(m *Context) {
	slog.Debug("Entering DiscordDiscordGoOnTextHandler")
	if m.Service == Discord {
		message := m.DiscordMessage
		if message != nil {
			m.rawText = message.Content
			m.id = message.ID
			if message.ReferencedMessage != nil {
				m.replyToId = message.ReferencedMessage.ID
				m.shouldReplyToMessage = true
			}
			m.chatId = message.ChannelID
		}
	}
	discordMessageParser.next.Execute(m)
}

func (mp *DiscordDiscordGoOnTextHandler) SetNext(next ContextHandler) {
	mp.next = next
}
