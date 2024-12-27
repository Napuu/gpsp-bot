package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/playwright-community/playwright-go"
)

type LatestEuriborRates struct {
	Date         time.Time
	ThreeMonths  float64
	SixMonths    float64
	TwelveMonths float64
}

const REPORTS_URL = "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"

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

func DownloadEuriborCSVFile(tmpFile *os.File) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		// For debugging
		// Headless: playwright.Bool(false),
	})
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

func GetEuriborRates() LatestEuriborRates {
	tmpFile, err := os.CreateTemp("", "euribor-*.csv")
	if err != nil {
		log.Fatalf("could not create temporary file: %v", err)
	}
	DownloadEuriborCSVFile(tmpFile)

	return ParseEuriborCSVFile(tmpFile.Name())
}
