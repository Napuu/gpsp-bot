package utils

import (
	"database/sql"
	"fmt"
	"time"
)

// DayVideoStateRow holds per-group day meme post schedule state.
type DayVideoStateRow struct {
	GroupID       string
	LastPostedAt  sql.NullTime
	LastCheckedAt sql.NullTime
	EligibleFrom  time.Time
	DueBy         time.Time
}

// GetDayVideoState returns schedule state for a group or sql.ErrNoRows when missing.
func GetDayVideoState(db *sql.DB, groupID string) (*DayVideoStateRow, error) {
	query := `
		SELECT group_id, last_posted_at, last_checked_at, eligible_from, due_by
		FROM day_video_state
		WHERE group_id = ?
	`
	row := db.QueryRow(query, groupID)

	var state DayVideoStateRow
	if err := row.Scan(&state.GroupID, &state.LastPostedAt, &state.LastCheckedAt, &state.EligibleFrom, &state.DueBy); err != nil {
		return nil, err
	}
	return &state, nil
}

// UpsertDayVideoState writes schedule state for a group.
func UpsertDayVideoState(db *sql.DB, state DayVideoStateRow) error {
	var lastPosted, lastChecked interface{}
	if state.LastPostedAt.Valid {
		lastPosted = state.LastPostedAt.Time
	}
	if state.LastCheckedAt.Valid {
		lastChecked = state.LastCheckedAt.Time
	}

	query := `
		INSERT INTO day_video_state (group_id, last_posted_at, last_checked_at, eligible_from, due_by)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (group_id) DO UPDATE SET
			last_posted_at = excluded.last_posted_at,
			last_checked_at = excluded.last_checked_at,
			eligible_from = excluded.eligible_from,
			due_by = excluded.due_by
	`
	_, err := db.Exec(query, state.GroupID, lastPosted, lastChecked, state.EligibleFrom, state.DueBy)
	if err != nil {
		return fmt.Errorf("upsert day video state: %w", err)
	}
	return nil
}

// ListActiveVideoGroups returns group IDs with enough recent non-repost
// group-chat bot video posts to be eligible for a day meme. Bot video activity
// is the sole signal that a group is alive; the background scan seeds schedule
// state and decides due-ness per group from this candidate set.
func ListActiveVideoGroups(db *sql.DB, now time.Time) ([]string, error) {
	videoSince := now.Add(-DayVideoVideoLookback)

	query := `
		SELECT group_id
		FROM video_stats
		WHERE is_repost = FALSE
		  AND is_group_chat = TRUE
		  AND posted_at >= ?
		GROUP BY group_id
		HAVING COUNT(*) >= ?
	`
	rows, err := db.Query(query, videoSince, dayVideoMinVideos7d)
	if err != nil {
		return nil, fmt.Errorf("list active video groups: %w", err)
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("scan active video group: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, rows.Err()
}

// CountRecentVideoPosts counts non-repost group-chat bot video posts since the given time.
func CountRecentVideoPosts(db *sql.DB, groupID string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM video_stats
		WHERE group_id = ? AND is_repost = FALSE AND is_group_chat = TRUE AND posted_at >= ?
	`
	var count int
	if err := db.QueryRow(query, groupID, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count recent video posts: %w", err)
	}
	return count, nil
}
