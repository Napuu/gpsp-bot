package utils

import (
	"fmt"
	"log"
	"os"

	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/playwright-community/playwright-go"
)

type Rate struct {
	Date  string
	Value float64
}

type LatestEuriborRates struct {
	ThreeMonths  Rate
	SixMonths    Rate
	TwelveMonths Rate
}

func ParseEuriborCSVFile(filePath string) LatestEuriborRates {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatalf("could not open DuckDB: %v", err)
	}
	defer conn.Close()

	query := fmt.Sprintf(`
		CREATE TABLE euribor AS 
		SELECT * FROM read_csv(
			'%s',
			header=false,
			skip=4,
			columns={'provider': 'VARCHAR', 'date': 'DATE', 'name': 'VARCHAR', 'value': 'VARCHAR'}
		);
		WITH latest_values AS (
			SELECT name, 
						 MAX(date) AS latest_date
			FROM euribor
			WHERE value IS NOT NULL
				AND name IN ('3 kk (tod.pv/360)', '6 kk (tod.pv/360)', '12 kk (tod.pv/360)')
			GROUP BY name
		)
		SELECT lv.name, 
					lv.latest_date, 
					CAST(REPLACE(e.value, ',', '.') AS DOUBLE) AS latest_value
		FROM latest_values lv
		JOIN euribor e 
			ON lv.name = e.name AND lv.latest_date = e.date
		ORDER BY lv.latest_date DESC;`, filePath)

	rows, err := conn.Query(query)
	if err != nil {
		log.Fatalf("could not query DuckDB: %v", err)
	}
	defer rows.Close()

	var rates LatestEuriborRates
	for rows.Next() {
		var name string
		var latestDate string
		var latestValue float64

		if err := rows.Scan(&name, &latestDate, &latestValue); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}

		rate := Rate{
			Date:  latestDate,
			Value: latestValue,
		}

		switch name {
		case "3 kk (tod.pv/360)":
			rates.ThreeMonths = rate
		case "6 kk (tod.pv/360)":
			rates.SixMonths = rate
		case "12 kk (tod.pv/360)":
			rates.TwelveMonths = rate
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
	suomepankkiReportsURL := "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"
	if _, err = page.Goto(suomepankkiReportsURL); err != nil {
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
