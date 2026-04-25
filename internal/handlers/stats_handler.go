package handlers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type StatsHandler struct {
	next ContextHandler
}

func (h *StatsHandler) Execute(m *Context) {
	slog.Debug("Entering StatsHandler")

	if m.action == Stats {
		m.disableWebPreview = true
		cfg := config.FromEnv()
		dbPath := filepath.Join(cfg.REPOST_DB_DIR, statsDBFileName)

		var platform string
		switch m.Service {
		case Telegram:
			platform = "telegram"
		case Discord:
			platform = "discord"
		default:
			slog.Warn("StatsHandler: unknown service", "service", m.Service)
			h.next.Execute(m)
			return
		}
		groupId := platform + ":" + m.chatId

		db, err := utils.OpenStatsDB(dbPath)
		if err != nil {
			slog.Warn("Failed to open stats DB", "error", err)
			m.textResponse = buildStatsText(m, nil, err, nil, err, nil, err, nil, err)
			h.next.Execute(m)
			return
		}
		defer db.Close()

		posters, postersErr := utils.GetGroupLeaderboard(db, groupId, 10)
		thumbsUp, thumbsUpErr := utils.GetTopThumbsUp(db, groupId, 5)
		thumbsDown, thumbsDownErr := utils.GetTopThumbsDown(db, groupId, 5)
		reposters, repostersErr := utils.GetTopReposters(db, groupId, 5)

		m.textResponse = buildStatsText(m, posters, postersErr, thumbsUp, thumbsUpErr, thumbsDown, thumbsDownErr, reposters, repostersErr)
	}

	h.next.Execute(m)
}

func (h *StatsHandler) SetNext(next ContextHandler) {
	h.next = next
}

func buildStatsText(
	m *Context,
	posters []utils.PosterStat, postersErr error,
	thumbsUp []utils.ReactionStat, thumbsUpErr error,
	thumbsDown []utils.ReactionStat, thumbsDownErr error,
	reposters []utils.RepostStat, repostersErr error,
) string {
	var sb strings.Builder

	sb.WriteString("Top video posters:\n")
	appendPosterSection(&sb, posters, postersErr)

	sb.WriteString("\n👍 Most liked:\n")
	appendReactionSection(m, &sb, thumbsUp, thumbsUpErr)

	sb.WriteString("\n👎 Most disliked:\n")
	appendReactionSection(m, &sb, thumbsDown, thumbsDownErr)

	sb.WriteString("\nTop reposters:\n")
	appendReposterSection(&sb, reposters, repostersErr)

	return sb.String()
}

func appendPosterSection(sb *strings.Builder, posters []utils.PosterStat, err error) {
	if err != nil {
		sb.WriteString("(error fetching data)\n")
		return
	}
	if len(posters) == 0 {
		sb.WriteString("No videos posted yet.\n")
		return
	}
	for i, p := range posters {
		name := p.Username
		if name == "" {
			name = p.UserId
		}
		sb.WriteString(fmt.Sprintf("%d. %s — %d video", i+1, name, p.PostCount))
		if p.PostCount != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n")
	}
}

func appendReactionSection(m *Context, sb *strings.Builder, videos []utils.ReactionStat, err error) {
	if err != nil {
		sb.WriteString("(error fetching data)\n")
		return
	}
	if len(videos) == 0 {
		sb.WriteString("None yet.\n")
		return
	}
	for i, v := range videos {
		name := v.Username
		if name == "" {
			name = "unknown"
		}
		
		url := v.SourceUrl
		if v.BotMessageId != "" {
			if m.Service == Telegram {
				if m.chatUsername != "" {
					url = fmt.Sprintf("https://t.me/%s/%s", m.chatUsername, v.BotMessageId)
				} else if strings.HasPrefix(m.chatId, "-100") {
					url = fmt.Sprintf("https://t.me/c/%s/%s", strings.TrimPrefix(m.chatId, "-100"), v.BotMessageId)
				} else {
					url = fmt.Sprintf("https://t.me/c/%s/%s", strings.TrimPrefix(m.chatId, "-"), v.BotMessageId)
				}
			} else if m.Service == Discord {
				guildId := m.guildId
				if guildId == "" {
					guildId = "@me"
				}
				url = fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildId, m.chatId, v.BotMessageId)
			}
		}

		sb.WriteString(fmt.Sprintf("%d. %s — %d (%s)\n", i+1, name, v.ReactionCount, url))
	}
}

func appendReposterSection(sb *strings.Builder, reposters []utils.RepostStat, err error) {
	if err != nil {
		sb.WriteString("(error fetching data)\n")
		return
	}
	if len(reposters) == 0 {
		sb.WriteString("No reposts detected.\n")
		return
	}
	for i, r := range reposters {
		name := r.Username
		if name == "" {
			name = r.UserId
		}
		sb.WriteString(fmt.Sprintf("%d. %s — %d repost", i+1, name, r.RepostCount))
		if r.RepostCount != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("\n")
	}
}
