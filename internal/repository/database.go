package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/pkg/utils"
)

type RateCache struct {
	LastFetched time.Time
	Value       utils.LatestEuriborRates
}

func InitializeDB() (*sql.DB, error) {
	databaseLocation := config.FromEnv().DATABASE_FILE
	db, err := sql.Open("sqlite3", databaseLocation)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS euribor_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			last_fetched DATETIME,
			date DATETIME,
			three_months REAL,
			six_months REAL,
			twelve_months REAL
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetCachedRates(db *sql.DB) (*RateCache, error) {
	cacheWindow := time.Minute * 30
	var rateCache RateCache

	err := db.QueryRow(`
		SELECT last_fetched, date, three_months, six_months, twelve_months
		FROM euribor_cache
		ORDER BY last_fetched DESC LIMIT 1`,
	).Scan(
		&rateCache.LastFetched,
		&rateCache.Value.Date,
		&rateCache.Value.ThreeMonths,
		&rateCache.Value.SixMonths,
		&rateCache.Value.TwelveMonths)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}
	latestDateIsOldEnough := time.Since(rateCache.Value.Date) > 24*time.Hour
	weekday := time.Now().Weekday()
	if weekday == time.Saturday {
		latestDateIsOldEnough = time.Since(rateCache.Value.Date) > 2*24*time.Hour
	}
	if weekday == time.Sunday {
		latestDateIsOldEnough = time.Since(rateCache.Value.Date) > 3*24*time.Hour
	}
	enoughTimeHasPassedSinceLastFetch := time.Since(rateCache.LastFetched) > cacheWindow

	if latestDateIsOldEnough && enoughTimeHasPassedSinceLastFetch {
		return nil, nil
	}

	return &rateCache, nil
}

func InsertRates(db *sql.DB, rates RateCache) error {
	_, err := db.Exec(`
		INSERT INTO euribor_cache (last_fetched, date, three_months, six_months, twelve_months)
		VALUES (?, ?, ?, ?, ?)`,
		rates.LastFetched, rates.Value.Date, rates.Value.ThreeMonths, rates.Value.SixMonths, rates.Value.TwelveMonths)
	if err != nil {
		return fmt.Errorf("failed to insert rates into cache: %w", err)
	}

	return nil
}
