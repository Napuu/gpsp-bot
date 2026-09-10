package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// GetRatesFromCSV must never crash the process when the csv directory does not
// exist yet. It should create the directory and return an empty result.
func TestGetRatesFromCSVCreatesMissingDir(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "euribor-exports")

	if _, err := os.Stat(missingDir); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist before the call", missingDir)
	}

	data := GetRatesFromCSV(missingDir, time.Now().AddDate(0, -1, 0))

	if len(data) != 0 {
		t.Fatalf("expected empty result for missing dir, got %d entries", len(data))
	}

	info, err := os.Stat(missingDir)
	if err != nil {
		t.Fatalf("expected directory to be created, stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", missingDir)
	}
}

func writeEuriborCSV(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
}

func assertEuriborRates(t *testing.T, data []EuriborRateEntry) {
	t.Helper()
	if len(data) != 1 {
		t.Fatalf("expected 1 date row, got %d", len(data))
	}
	got := data[0]
	if got.Date.Format("2006-01-02") != "2026-09-03" {
		t.Fatalf("unexpected date: %v", got.Date)
	}
	if got.ThreeMonths != 2.655 {
		t.Fatalf("unexpected 3m rate: %v", got.ThreeMonths)
	}
	if got.SixMonths != 2.789 {
		t.Fatalf("unexpected 6m rate: %v", got.SixMonths)
	}
	if got.TwelveMonths != 3.109 {
		t.Fatalf("unexpected 12m rate: %v", got.TwelveMonths)
	}
}

func TestGetRatesFromCSVWithBlankLineAndValueHeaders(t *testing.T) {
	dir := t.TempDir()
	writeEuriborCSV(t, dir, "new.csv", ""+
		"title,description,creator,frequency,date_format,updated\r\n"+
		"Euriborkorot ja Eoniakorko,\"desc\",Suomen Pankki,D,yyyy-MM-dd,2026-09-03\r\n"+
		"\r\n"+
		"provider,value,name,value\r\n"+
		"Reuters,2026-09-03,Eonia (tod.pv/360),\r\n"+
		"Reuters,2026-09-03,3 kk (tod.pv/360),\"2,655\"\r\n"+
		"Reuters,2026-09-03,6 kk (tod.pv/360),\"2,789\"\r\n"+
		"Reuters,2026-09-03,12 kk (tod.pv/360),\"3,109\"\r\n")

	assertEuriborRates(t, GetRatesFromCSV(dir, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))
}

func TestGetRatesFromCSVWithLegacyDateRateHeaders(t *testing.T) {
	dir := t.TempDir()
	writeEuriborCSV(t, dir, "legacy.csv", ""+
		"title,description,creator,frequency,date_format,updated\r\n"+
		"Euriborkorot ja Eoniakorko,\"desc\",Suomen Pankki,D,yyyy-MM-dd,2026-09-03\r\n"+
		"provider,date,name,rate\r\n"+
		"Reuters,2026-09-03,3 kk (tod.pv/360),\"2,655\"\r\n"+
		"Reuters,2026-09-03,6 kk (tod.pv/360),\"2,789\"\r\n"+
		"Reuters,2026-09-03,12 kk (tod.pv/360),\"3,109\"\r\n")

	assertEuriborRates(t, GetRatesFromCSV(dir, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))
}

func TestGetRatesFromCSVWithExtraBlankLines(t *testing.T) {
	dir := t.TempDir()
	writeEuriborCSV(t, dir, "messy.csv", ""+
		"title,description,creator,frequency,date_format,updated\r\n"+
		"Euriborkorot ja Eoniakorko,\"desc\",Suomen Pankki,D,yyyy-MM-dd,2026-09-03\r\n"+
		"\r\n"+
		"\r\n"+
		"provider,value,name,value\r\n"+
		"\r\n"+
		"Reuters,2026-09-03,3 kk (tod.pv/360),\"2,655\"\r\n"+
		"Reuters,2026-09-03,6 kk (tod.pv/360),\"2,789\"\r\n"+
		"Reuters,2026-09-03,12 kk (tod.pv/360),\"3,109\"\r\n")

	assertEuriborRates(t, GetRatesFromCSV(dir, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))
}
