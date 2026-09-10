package utils

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/snapshot-chromedp/render"
	"github.com/mxschmitt/playwright-go"
)

const REPORTS_URL = "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"

type EuriborRateEntry struct {
	Date         time.Time
	ThreeMonths  float64
	SixMonths    float64
	TwelveMonths float64
}

func GenerateLine(data []EuriborRateEntry, outputPath string) error {
	slog.Debug("Generating line")
	line := charts.NewLine()

	// Get unique dates and build data series
	dateMap := make(map[string]map[string]float64)
	for _, entry := range data {
		dateStr := entry.Date.Format("2006-01-02")
		if _, exists := dateMap[dateStr]; !exists {
			dateMap[dateStr] = map[string]float64{}
		}
		dateMap[dateStr]["3 kk (tod.pv/360)"] = entry.ThreeMonths
		dateMap[dateStr]["6 kk (tod.pv/360)"] = entry.SixMonths
		dateMap[dateStr]["12 kk (tod.pv/360)"] = entry.TwelveMonths
	}

	// Sort dates
	var dates []string
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Find the min and max rate values
	var minRate, maxRate float64
	for i, name := range []string{"3 kk (tod.pv/360)", "6 kk (tod.pv/360)", "12 kk (tod.pv/360)"} {
		for _, date := range dates {
			val := dateMap[date][name]
			if i == 0 && date == dates[0] { // Initialize min and max with the first rate
				minRate = val
				maxRate = val
			}
			if val < minRate {
				minRate = val
			}
			if val > maxRate {
				maxRate = val
			}
		}
	}

	// Set global options for the chart
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			BackgroundColor: "#FFFFFF",
		}),
		charts.WithAnimation(false),
		charts.WithTitleOpts(opts.Title{Title: "Euribor Rates - Last 30 Days"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Date"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Rate (%)", Max: fmt.Sprintf("%.1f", maxRate+0.1), Min: fmt.Sprintf("%.1f", minRate-0.1)}),
	)

	// Add axis data (dates)
	line.SetXAxis(dates)

	// Add each rate type as a series (line)
	for _, name := range []string{"3 kk (tod.pv/360)", "6 kk (tod.pv/360)", "12 kk (tod.pv/360)"} {
		var values []opts.LineData
		for _, date := range dates {
			val := dateMap[date][name]
			values = append(values, opts.LineData{Value: val})
		}
		line.AddSeries(name, values)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory for chart: %v", err)
	}

	err := render.MakeChartSnapshot(line.RenderContent(), outputPath)
	if err != nil {
		return fmt.Errorf("failed to render the line chart: %v", err)
	}

	// Verify the output file was actually created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("chart rendering completed without error, but output file was not created at %s", outputPath)
	}

	return nil
}

func DownloadEuriborCSVFile(filePath string) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(REPORTS_URL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	if err = page.Locator("a[title=Export]").Click(); err != nil {
		log.Fatalf("could not click the Export button: %v", err)
	}
	download, err := page.ExpectDownload(func() error {
		return page.Locator("text=CSV (comma delimited").Click()
	})
	if err != nil {
		log.Fatalf("could not trigger download: %v", err)
	}
	if err = download.SaveAs(filePath); err != nil {
		log.Fatalf("could not save file: %v", err)
	}
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func isLatestCSVOlderThan(dirPath string, maxAge time.Duration) (bool, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	var latestFile os.FileInfo
	for _, file := range files {
		if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".csv") {
			info, err := file.Info()
			if err != nil {
				return false, err
			}
			if latestFile == nil || info.ModTime().After(latestFile.ModTime()) {
				latestFile = info
			}
		}
	}

	if latestFile == nil {
		return false, fmt.Errorf("no csv files found in %s", dirPath)
	}

	maxAgeAgo := time.Now().Add(-maxAge)
	return latestFile.ModTime().Before(maxAgeAgo), nil
}

func ShouldFetchCSV(dirPath string, maxAge time.Duration) bool {
	history := GetRatesFromCSV(dirPath, time.Now().AddDate(0, -1, 0))

	today := time.Now()
	for _, entry := range history {
		if entry.Date.Year() == today.Year() && entry.Date.Month() == today.Month() && entry.Date.Day() == today.Day() {
			return false
		}
	}

	latestFileStale, err := isLatestCSVOlderThan(dirPath, maxAge)
	if err != nil {
		return true
	}

	if latestFileStale {
		return true
	}

	return false
}

func GetRatesFromCSV(filePath string, startDate time.Time) []EuriborRateEntry {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatalf("could not open DuckDB: %v", err)
	}
	defer conn.Close()

	if err := os.MkdirAll(filePath, 0755); err != nil {
		slog.Warn("could not create euribor csv directory", "path", filePath, "error", err)
		return []EuriborRateEntry{}
	}

	files, err := os.ReadDir(filePath)
	if err != nil {
		slog.Warn("could not read euribor csv directory", "path", filePath, "error", err)
		return []EuriborRateEntry{}
	}

	if len(files) == 0 {
		return []EuriborRateEntry{}
	}

	// Read every row as VARCHAR and keep only rows whose 2nd field is a date.
	// That ignores Bank of Finland metadata, blank lines, and renamed headers
	// (e.g. date/rate vs value/value) without relying on a fixed SKIP.
	query := `
		WITH
		  raw_interest_rates AS (
			SELECT *
			FROM read_csv("` + strings.TrimSuffix(filePath, "/") + `/*",
			  header = false,
			  delim = ',',
			  quote = '"',
			  all_varchar = true,
			  ignore_errors = true,
			  null_padding = true
			)
		  ),
		  interest_rates AS (
			SELECT
			  column0 AS provider,
			  TRY_CAST(column1 AS DATE) AS date,
			  column2 AS name,
			  TRY_CAST(REPLACE(column3, ',', '.') AS DOUBLE) AS rate
			FROM raw_interest_rates
			WHERE TRY_CAST(column1 AS DATE) IS NOT NULL
		  )
		SELECT
			date,
			MAX(CASE WHEN name = '3 kk (tod.pv/360)' THEN rate END) AS threemonths,
			MAX(CASE WHEN name = '6 kk (tod.pv/360)' THEN rate END) AS sixmonths,
			MAX(CASE WHEN name = '12 kk (tod.pv/360)' THEN rate END) AS twelvemonths
		FROM interest_rates
		WHERE
			rate IS NOT NULL AND
			date >= '` + startDate.Format("2006-01-02") + `'
		GROUP BY date
		ORDER BY date DESC;
	`

	rows, err := conn.Query(query)
	if err != nil {
		slog.Warn("could not query DuckDB for euribor csv", "error", err)
		return []EuriborRateEntry{}
	}
	defer rows.Close()

	var history []EuriborRateEntry
	for rows.Next() {
		var entry EuriborRateEntry
		if err := rows.Scan(&entry.Date, &entry.ThreeMonths, &entry.SixMonths, &entry.TwelveMonths); err != nil {
			slog.Warn("could not scan euribor row", "error", err)
			return []EuriborRateEntry{}
		}
		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		slog.Warn("error iterating euribor rows", "error", err)
		return []EuriborRateEntry{}
	}

	return history
}
