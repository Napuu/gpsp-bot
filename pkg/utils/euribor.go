package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/snapshot-chromedp/render"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/playwright-community/playwright-go"
)

const REPORTS_URL = "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"

type LatestEuriborRates struct {
	Date         time.Time
	ThreeMonths  float64
	SixMonths    float64
	TwelveMonths float64
}

type EuriborRateEntry struct {
	Date time.Time
	Name string
	Rate float64
}

type EuriborData struct {
	Latest   LatestEuriborRates
	History  []EuriborRateEntry
	FilePath string
}

func GetEuriborData() EuriborData {
	tmpFile, err := os.CreateTemp("", "euribor-*.csv")
	if err != nil {
		log.Fatalf("could not create temporary file: %v", err)
	}

	DownloadEuriborCSVFile(tmpFile)

	return EuriborData{
		Latest:   ParseEuriborCSVFile(tmpFile.Name()),
		History:  ParseEuriborHistory(tmpFile.Name()),
		FilePath: tmpFile.Name(),
	}
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
		dateMap[dateStr][entry.Name] = entry.Rate
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
		charts.WithYAxisOpts(opts.YAxis{Name: "Rate (%)", Max: maxRate + 0.1, Min: minRate - 0.1}),
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

func DownloadEuriborCSVFile(tmpFile *os.File) {
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
	if err = download.SaveAs(tmpFile.Name()); err != nil {
		log.Fatalf("could not save file: %v", err)
	}
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func ParseEuriborCSVFile(filePath string) LatestEuriborRates {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatalf("could not open DuckDB: %v", err)
	}
	defer conn.Close()

	query := fmt.Sprintf(`
		WITH data AS (
			SELECT *
			FROM read_csv(
				'%s',
				header=false,
				skip=4,
				columns={'provider': 'VARCHAR', 'date': 'DATE', 'name': 'VARCHAR', 'value': 'VARCHAR'}
			)
			WHERE value IS NOT NULL
				AND name IN ('3 kk (tod.pv/360)', '6 kk (tod.pv/360)', '12 kk (tod.pv/360)')
		),
		latest_date AS (
			SELECT MAX(date) AS latest_date
			FROM data
		),
		latest_values AS (
			SELECT
				MAX(CASE WHEN name = '3 kk (tod.pv/360)' THEN CAST(REPLACE(value, ',', '.') AS DOUBLE) END) AS three_months_rate,
				MAX(CASE WHEN name = '6 kk (tod.pv/360)' THEN CAST(REPLACE(value, ',', '.') AS DOUBLE) END) AS six_months_rate,
				MAX(CASE WHEN name = '12 kk (tod.pv/360)' THEN CAST(REPLACE(value, ',', '.') AS DOUBLE) END) AS twelve_months_rate
			FROM data
			JOIN latest_date ON data.date = latest_date.latest_date
		)
		SELECT 
			latest_date.latest_date,
			latest_values.three_months_rate,
			latest_values.six_months_rate,
			latest_values.twelve_months_rate
		FROM latest_date
		CROSS JOIN latest_values;`, filePath)

	rows, err := conn.Query(query)
	if err != nil {
		log.Fatalf("could not query DuckDB: %v", err)
	}
	defer rows.Close()

	var rates LatestEuriborRates
	for rows.Next() {
		if err := rows.Scan(&rates.Date, &rates.ThreeMonths, &rates.SixMonths, &rates.TwelveMonths); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("error iterating rows: %v", err)
	}

	return rates
}

func ParseEuriborHistory(filePath string) []EuriborRateEntry {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatalf("could not open DuckDB: %v", err)
	}
	defer conn.Close()

	query := fmt.Sprintf(`
		WITH data AS (
			SELECT *
			FROM read_csv(
				'%s',
				header=false,
				skip=4,
				columns={'provider': 'VARCHAR', 'date': 'DATE', 'name': 'VARCHAR', 'value': 'VARCHAR'}
			)
			WHERE value IS NOT NULL
				AND name IN ('3 kk (tod.pv/360)', '6 kk (tod.pv/360)', '12 kk (tod.pv/360)')
		)
		SELECT 
			date,
			name,
			CAST(REPLACE(value, ',', '.') AS DOUBLE) AS rate
		FROM data
		WHERE date >= CURRENT_DATE - INTERVAL 30 DAY
		ORDER BY date;
	`, filePath)

	rows, err := conn.Query(query)
	if err != nil {
		log.Fatalf("could not query DuckDB: %v", err)
	}
	defer rows.Close()

	var history []EuriborRateEntry
	for rows.Next() {
		var entry EuriborRateEntry
		if err := rows.Scan(&entry.Date, &entry.Name, &entry.Rate); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("error iterating rows: %v", err)
	}

	return history
}
