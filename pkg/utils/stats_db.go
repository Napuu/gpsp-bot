package utils

import (
	"database/sql"
	"fmt"
	"slices"
	"time"
)

var positiveEmojis = []string{"👍", "❤️", "🔥", "🥰", "👏", "😁", "🤩", "🎉", "💯", "✅"}
var negativeEmojis = []string{"👎", "💩", "😡", "🤮"}

// VideoStatEntry holds the data for one recorded video post.
type VideoStatEntry struct {
	Platform     string
	GroupId      string
	UserId       string
	Username     string
	SourceUrl    string
	BotMessageId string
	IsRepost     bool
	PostedAt     time.Time
}

// PosterStat holds aggregated posting stats for one user.
type PosterStat struct {
	UserId    string
	Username  string
	PostCount int
}

// ReactionStat holds the most-reacted video entries.
type ReactionStat struct {
	BotMessageId  string
	Username      string
	SourceUrl     string
	ReactionCount int
	PostedAt      time.Time
}

// RepostStat holds repost counts per user.
type RepostStat struct {
	UserId      string
	Username    string
	RepostCount int
}

// OpenStatsDB opens a DuckDB connection for stats queries. Callers are
// responsible for closing the returned *sql.DB.
func OpenStatsDB(dbPath string) (*sql.DB, error) {
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return conn, nil
}

// RecordVideoPost inserts one row into video_stats for a successful video send.
func RecordVideoPost(db *sql.DB, entry VideoStatEntry) error {
	query := `
		INSERT INTO video_stats (platform, group_id, user_id, username, source_url, bot_message_id, is_repost, posted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query,
		entry.Platform, entry.GroupId, entry.UserId, entry.Username,
		entry.SourceUrl, entry.BotMessageId, entry.IsRepost, entry.PostedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record video stat: %w", err)
	}
	return nil
}

// UpdateReactionCount adjusts per-emoji reaction counts for the given bot
// message. emoji should be the raw emoji character (e.g. "👍"). Unrecognised
// emoji are ignored. A delta that matches no row is silently ignored.
func UpdateReactionCount(db *sql.DB, platform, groupId, botMessageId, emoji string, delta int) error {
	var column string
	switch {
	case slices.Contains(positiveEmojis, emoji):
		column = "thumbs_up_count"
	case slices.Contains(negativeEmojis, emoji):
		column = "thumbs_down_count"
	default:
		return nil // unrecognised emoji, nothing to update
	}

	query := fmt.Sprintf(`
		UPDATE video_stats
		SET %s = %s + ?
		WHERE platform = ? AND group_id = ? AND bot_message_id = ?
	`, column, column)
	_, err := db.Exec(query, delta, platform, groupId, botMessageId)
	if err != nil {
		return fmt.Errorf("failed to update reaction count: %w", err)
	}
	return nil
}

// GetGroupLeaderboard returns the top N posters by post count for the given group.
// Groups strictly by user_id so that a username change does not split a user.
func GetGroupLeaderboard(db *sql.DB, groupId string, limit int) ([]PosterStat, error) {
	query := `
		SELECT user_id, MAX(username) AS username, COUNT(*) AS post_count
		FROM video_stats
		WHERE group_id = ? AND is_repost = FALSE
		GROUP BY user_id
		ORDER BY post_count DESC
		LIMIT ?
	`
	rows, err := db.Query(query, groupId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var result []PosterStat
	for rows.Next() {
		var s PosterStat
		if err := rows.Scan(&s.UserId, &s.Username, &s.PostCount); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetTopThumbsUp returns the top N videos by 👍 count for the given group.
func GetTopThumbsUp(db *sql.DB, groupId string, limit int) ([]ReactionStat, error) {
	return queryTopByReactionColumn(db, groupId, "thumbs_up_count", limit)
}

// GetTopThumbsDown returns the top N videos by 👎 count for the given group.
func GetTopThumbsDown(db *sql.DB, groupId string, limit int) ([]ReactionStat, error) {
	return queryTopByReactionColumn(db, groupId, "thumbs_down_count", limit)
}

func queryTopByReactionColumn(db *sql.DB, groupId, column string, limit int) ([]ReactionStat, error) {
	if column != "thumbs_up_count" && column != "thumbs_down_count" {
		return nil, fmt.Errorf("invalid reaction column: %s", column)
	}

	query := fmt.Sprintf(`
		SELECT bot_message_id, username, source_url, COALESCE(%s, 0), posted_at
		FROM video_stats
		WHERE group_id = ? AND COALESCE(%s, 0) > 0
		ORDER BY COALESCE(%s, 0) DESC
		LIMIT ?
	`, column, column, column)

	rows, err := db.Query(query, groupId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top %s: %w", column, err)
	}
	defer rows.Close()

	var result []ReactionStat
	for rows.Next() {
		var s ReactionStat
		if err := rows.Scan(&s.BotMessageId, &s.Username, &s.SourceUrl, &s.ReactionCount, &s.PostedAt); err != nil {
			return nil, fmt.Errorf("failed to scan reaction stat row: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetTopReposters returns the top N users who have been caught reposting in the given group.
func GetTopReposters(db *sql.DB, groupId string, limit int) ([]RepostStat, error) {
	query := `
		SELECT user_id, MAX(username) AS username, COUNT(*) AS repost_count
		FROM video_stats
		WHERE group_id = ? AND is_repost = TRUE
		GROUP BY user_id
		ORDER BY repost_count DESC
		LIMIT ?
	`
	rows, err := db.Query(query, groupId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top reposters: %w", err)
	}
	defer rows.Close()

	var result []RepostStat
	for rows.Next() {
		var s RepostStat
		if err := rows.Scan(&s.UserId, &s.Username, &s.RepostCount); err != nil {
			return nil, fmt.Errorf("failed to scan repost stat row: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
