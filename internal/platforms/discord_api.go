package platforms

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
)

func wrapDiscoHandler(chain *chain.HandlerChain) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Wrap the context
		var replyToId string
		isReply := m.ReferencedMessage != nil
		if isReply {
			replyToId = m.ReferencedMessage.ID
		}

		chain.Process(&handlers.Context{
			DiscordSession:      s,
			DiscordMessage:      m,
			Service:             handlers.Discord,
			RawText:             m.Content,
			ID:                  m.ID,
			ChatID:              m.ChannelID,
			IsReply:             isReply,
			ReplyToID:           replyToId,
			ShouldReplyToMessage: isReply,
		})
	}
}

func RunDiscordBot() {
	token := config.FromEnv().DISCORD_TOKEN
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("Error creating Discord session", "error", err)
		return
	}

	// Create the chain of responsibility
	chain := chain.NewChainOfResponsibility()

	// Add a handler for messages
	dg.AddHandler(wrapDiscoHandler(chain))

	// Specify intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Open the connection
	err = dg.Open()
	if err != nil {
		slog.Error("Error opening Discord connection", "error", err)
		return
	}
}
