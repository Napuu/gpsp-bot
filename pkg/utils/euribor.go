package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	fmt.Println("pingpong")
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	fmt.Println("continuing")
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	fmt.Println("opening page")
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	fmt.Println("navigating")
	suomepankkiReportsURL := "https://reports.suomenpankki.fi/WebForms/ReportViewerPage.aspx?report=/tilastot/markkina-_ja_hallinnolliset_korot/euribor_korot_xml_long_fi&output=html"
	if _, err = page.Goto(suomepankkiReportsURL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	fmt.Println("locating the button with text=Export and clicking it")
	if err = page.Locator("a[title=Export]").Click(); err != nil {
		log.Fatalf("could not click the Export button: %v", err)
	}

	fmt.Println("trying to download next")
	download, err := page.ExpectDownload(func() error {
		return page.Locator("text=CSV (comma delimited").Click()
	})

	fmt.Println(download.SuggestedFilename())

	tmpDir, _ := os.CreateTemp("", "euribor")

	defer f.Close()
	defer os.Remove(f.Name())

	download.SaveAs(tmpDir.Name())
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	reader := csv.NewReader(tmpDir)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	targets := map[string]string{
		"3 kk (tod.pv/360)":  "",
		"6 kk (tod.pv/360)":  "",
		"12 kk (tod.pv/360)": "",
	}

	for _, row := range records {
		// discard first 3 rows, they contain metadata etc
		if len(row) < 4 {
			continue
		}

		name := strings.TrimSpace(row[2])
		value := strings.TrimSpace(row[3])

		if v, ok := targets[name]; ok && v == "" {
			targets[name] = value
		}
	}

	for name, value := range targets {
		fmt.Printf("%s: %s\n", name, value)
	}

}
