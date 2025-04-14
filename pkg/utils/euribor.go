package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/snapshot-chromedp/render"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/playwright-community/playwright-go"
)

const REPORTS_URL = "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"

type EuriborRateEntry struct {
	Date         time.Time
	ThreeMonths  float64
	SixMonths    float64
	TwelveMonths float64
}

func GenerateLine(data []EuriborRateEntry, outputPath string) error {
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

	err := render.MakeChartSnapshot(line.RenderContent(), outputPath)
	if err != nil {
		return fmt.Errorf("failed to render the line chart: %v", err)
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

	files, err := os.ReadDir(filePath)
	if err != nil {
		log.Fatalf("could not read directory: %v", err)
	}

	if len(files) == 0 {
		return []EuriborRateEntry{}
	}

	query := `
		WITH
		  raw_interest_rates AS (
			SELECT *
			FROM read_csv_auto("` + strings.TrimSuffix(filePath, "/") + `/*",
			  HEADER=false,
			  DELIM=',',
			  QUOTE='"',
			  SKIP=3,
			  COLUMNS={'provider': 'VARCHAR', 'date': 'DATE', 'name': 'VARCHAR', 'rate': 'VARCHAR'}
			)
		  ),
		  interest_rates AS (
			SELECT provider,
			CAST(date AS DATE) as date,
			name,
			CAST(REPLACE(rate, ',', '.') AS DOUBLE) AS rate
			FROM raw_interest_rates
		  )
		SELECT
			date,
			MAX(CASE WHEN name = '3 kk (tod.pv/360)' THEN rate END) AS threemonths,
			MAX(CASE WHEN name = '6 kk (tod.pv/360)' THEN rate END) AS sixmonths,
			MAX(CASE WHEN name = '12 kk (tod.pv/360)' THEN rate END) AS twelvemonths,
		FROM interest_rates
		WHERE
			rate IS NOT NULL AND
			date >= '` + startDate.Format("2006-01-02") + `' GROUP BY date ORDER BY DATE DESC;
	`

	rows, err := conn.Query(query)
	if err != nil {
		log.Fatalf("could not query DuckDB: %v", err)
	}
	defer rows.Close()

	var history []EuriborRateEntry
	for rows.Next() {
		var entry EuriborRateEntry
		if err := rows.Scan(&entry.Date, &entry.ThreeMonths, &entry.SixMonths, &entry.TwelveMonths); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("error iterating rows: %v", err)
	}

	return history
}
