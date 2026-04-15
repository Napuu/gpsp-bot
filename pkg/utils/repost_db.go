package utils

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

const (
	repostDBFileName = "repost_fingerprints.duckdb"
)

// InitRepostDB creates the database file and schema if needed
func InitRepostDB(dbPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open connection to persistent DuckDB file
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer conn.Close()

	// Create schema
	// Use a sequence for auto-incrementing id
	schema := `
		CREATE SEQUENCE IF NOT EXISTS fingerprint_id_seq;
		CREATE TABLE IF NOT EXISTS fingerprints (
			id BIGINT PRIMARY KEY DEFAULT nextval('fingerprint_id_seq'),
			fingerprint BLOB NOT NULL,
			group_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
		CREATE SEQUENCE IF NOT EXISTS video_stats_id_seq;
		CREATE TABLE IF NOT EXISTS video_stats (
			id            BIGINT PRIMARY KEY DEFAULT nextval('video_stats_id_seq'),
			platform      TEXT NOT NULL,
			group_id      TEXT NOT NULL,
			user_id       TEXT NOT NULL,
			username      TEXT NOT NULL,
			source_url    TEXT NOT NULL,
			bot_message_id TEXT NOT NULL,
			reaction_count  INT NOT NULL DEFAULT 0,
			thumbs_up_count INT NOT NULL DEFAULT 0,
			thumbs_down_count INT NOT NULL DEFAULT 0,
			is_repost     BOOLEAN NOT NULL DEFAULT FALSE,
			posted_at     TIMESTAMP NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_video_stats_lookup
			ON video_stats (platform, group_id, bot_message_id);
		ALTER TABLE video_stats ADD COLUMN IF NOT EXISTS thumbs_up_count INT DEFAULT 0;
		ALTER TABLE video_stats ADD COLUMN IF NOT EXISTS thumbs_down_count INT DEFAULT 0;
		ALTER TABLE video_stats ADD COLUMN IF NOT EXISTS is_repost BOOLEAN DEFAULT FALSE;
	`

	_, err = conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// StoreFingerprint inserts a new fingerprint with group_id, message_id and timestamp
func StoreFingerprint(dbPath string, fingerprint []byte, groupId string, messageId string) error {
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer conn.Close()

	query := `
		INSERT INTO fingerprints (fingerprint, group_id, message_id, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err = conn.Exec(query, fingerprint, groupId, messageId, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store fingerprint: %w", err)
	}

	return nil
}

// FindSimilarFingerprints queries fingerprints for the specific group_id,
// compares using CalculateSimilarity, and returns message IDs of matches above threshold
func FindSimilarFingerprints(dbPath string, fingerprint []byte, groupId string, threshold float64) ([]string, error) {
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer conn.Close()

	// Query all fingerprints for this group_id
	query := `
		SELECT fingerprint, message_id
		FROM fingerprints
		WHERE group_id = ?
	`

	rows, err := conn.Query(query, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to query fingerprints: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var storedFingerprint []byte
		var messageId string

		if err := rows.Scan(&storedFingerprint, &messageId); err != nil {
			slog.Warn("Failed to scan fingerprint row", "error", err)
			continue
		}

		similarity := CalculateSimilarity(fingerprint, storedFingerprint)
		if similarity >= threshold {
			matches = append(matches, messageId)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return matches, nil
}

// CleanupOldFingerprints removes entries older than maxAge across all groups
func CleanupOldFingerprints(dbPath string, maxAge time.Duration) error {
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer conn.Close()

	cutoffTime := time.Now().Add(-maxAge)

	query := `
		DELETE FROM fingerprints
		WHERE created_at < ?
	`

	result, err := conn.Exec(query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old fingerprints: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		slog.Debug("Cleaned up old fingerprints", "count", rowsAffected)
	}

	return nil
}
