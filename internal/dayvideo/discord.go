package dayvideo

import (
	"github.com/bwmarrin/discordgo"
)

// DiscordMemberCount returns guild member count from session cache or REST fallback.
func DiscordMemberCount(session *discordgo.Session, guildID string) int {
	if session == nil || guildID == "" {
		return 0
	}
	if guild, err := session.State.Guild(guildID); err == nil && guild != nil && guild.MemberCount > 0 {
		return guild.MemberCount
	}
	guild, err := session.GuildWithCounts(guildID)
	if err != nil || guild == nil {
		return 0
	}
	return guild.MemberCount
}
