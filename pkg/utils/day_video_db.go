package utils

import (
	"database/sql"
	"fmt"
	"time"
)

// GroupActivityRow holds persisted group activity for day meme scheduling.
type GroupActivityRow struct {
	GroupID              string
	Platform             string
	LastMessageAt        time.Time
	MemberCount          sql.NullInt64
	MemberCountUpdatedAt sql.NullTime
}

// DayVideoStateRow holds per-group day meme post schedule state.
type DayVideoStateRow struct {
	GroupID      string
	LastPostedAt sql.NullTime
	EligibleFrom time.Time
	DueBy        time.Time
}

// RecordGroupActivity upserts last message time and optionally member count.
func RecordGroupActivity(db *sql.DB, groupID, platform string, memberCount *int, at time.Time) error {
	if memberCount != nil {
		query := `
			INSERT INTO group_activity (group_id, platform, last_message_at, member_count, member_count_updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (group_id) DO UPDATE SET
				platform = excluded.platform,
				last_message_at = excluded.last_message_at,
				member_count = excluded.member_count,
				member_count_updated_at = excluded.member_count_updated_at
		`
		_, err := db.Exec(query, groupID, platform, at, *memberCount, at)
		if err != nil {
			return fmt.Errorf("record group activity: %w", err)
		}
		return nil
	}

	query := `
		INSERT INTO group_activity (group_id, platform, last_message_at)
		VALUES (?, ?, ?)
		ON CONFLICT (group_id) DO UPDATE SET
			platform = excluded.platform,
			last_message_at = excluded.last_message_at
	`
	_, err := db.Exec(query, groupID, platform, at)
	if err != nil {
		return fmt.Errorf("record group activity: %w", err)
	}
	return nil
}

// GetGroupActivity returns activity for a group or sql.ErrNoRows when missing.
func GetGroupActivity(db *sql.DB, groupID string) (*GroupActivityRow, error) {
	query := `
		SELECT group_id, platform, last_message_at, member_count, member_count_updated_at
		FROM group_activity
		WHERE group_id = ?
	`
	row := db.QueryRow(query, groupID)

	var activity GroupActivityRow
	if err := row.Scan(
		&activity.GroupID,
		&activity.Platform,
		&activity.LastMessageAt,
		&activity.MemberCount,
		&activity.MemberCountUpdatedAt,
	); err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetDayVideoState returns schedule state for a group or sql.ErrNoRows when missing.
func GetDayVideoState(db *sql.DB, groupID string) (*DayVideoStateRow, error) {
	query := `
		SELECT group_id, last_posted_at, eligible_from, due_by
		FROM day_video_state
		WHERE group_id = ?
	`
	row := db.QueryRow(query, groupID)

	var state DayVideoStateRow
	if err := row.Scan(&state.GroupID, &state.LastPostedAt, &state.EligibleFrom, &state.DueBy); err != nil {
		return nil, err
	}
	return &state, nil
}

// UpsertDayVideoState writes schedule state for a group.
func UpsertDayVideoState(db *sql.DB, state DayVideoStateRow) error {
	var lastPosted interface{}
	if state.LastPostedAt.Valid {
		lastPosted = state.LastPostedAt.Time
	}

	query := `
		INSERT INTO day_video_state (group_id, last_posted_at, eligible_from, due_by)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (group_id) DO UPDATE SET
			last_posted_at = excluded.last_posted_at,
			eligible_from = excluded.eligible_from,
			due_by = excluded.due_by
	`
	_, err := db.Exec(query, state.GroupID, lastPosted, state.EligibleFrom, state.DueBy)
	if err != nil {
		return fmt.Errorf("upsert day video state: %w", err)
	}
	return nil
}

// ListDueGroups returns group IDs where now >= eligible_from.
func ListDueGroups(db *sql.DB, now time.Time) ([]string, error) {
	query := `
		SELECT group_id
		FROM day_video_state
		WHERE eligible_from <= ?
	`
	rows, err := db.Query(query, now)
	if err != nil {
		return nil, fmt.Errorf("list due groups: %w", err)
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("scan due group: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, rows.Err()
}

// ListDueActiveGroups returns due group IDs that pass activity gates for background scans.
func ListDueActiveGroups(db *sql.DB, now time.Time) ([]string, error) {
	activitySince := now.Add(-dayVideoActivityWindow)
	videoSince := now.Add(-DayVideoVideoLookback)

	query := `
		SELECT s.group_id
		FROM day_video_state s
		INNER JOIN group_activity a ON a.group_id = s.group_id
		WHERE s.eligible_from <= ?
		  AND a.member_count >= ?
		  AND a.last_message_at >= ?
		  AND (
			SELECT COUNT(*)
			FROM video_stats v
			WHERE v.group_id = s.group_id
			  AND v.is_repost = FALSE
			  AND v.posted_at >= ?
		  ) >= ?
	`
	rows, err := db.Query(query, now, dayVideoMinMembers, activitySince, videoSince, dayVideoMinVideos7d)
	if err != nil {
		return nil, fmt.Errorf("list due active groups: %w", err)
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("scan due active group: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, rows.Err()
}

// CountRecentVideoPosts counts non-repost bot video posts since the given time.
func CountRecentVideoPosts(db *sql.DB, groupID string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM video_stats
		WHERE group_id = ? AND is_repost = FALSE AND posted_at >= ?
	`
	var count int
	if err := db.QueryRow(query, groupID, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count recent video posts: %w", err)
	}
	return count, nil
}
