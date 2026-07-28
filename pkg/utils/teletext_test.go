package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var samplePNG = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x01, 0x02}

func teletextResponseBody(t *testing.T, mimeType string, image []byte) string {
	t.Helper()
	img := fmt.Sprintf(`<img src="data:%s;base64,%s" alt="" />`, mimeType, base64.StdEncoding.EncodeToString(image))
	body, err := json.Marshal(map[string]any{
		"meta": map[string]string{"code": "200"},
		"data": []map[string]any{
			{"content": map[string]string{"image": img}},
		},
	})
	if err != nil {
		t.Fatalf("failed to build response body: %v", err)
	}
	return string(body)
}

func TestFetchTeletextPage(t *testing.T) {
	var requestedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedQuery = r.URL.RawQuery
		fmt.Fprint(w, teletextResponseBody(t, "image/png", samplePNG))
	}))
	defer server.Close()

	page, err := fetchTeletextPage(server.URL, "867_0002")
	if err != nil {
		t.Fatalf("fetchTeletextPage() error: %v", err)
	}

	if requestedQuery != "P=867_0002" {
		t.Errorf("query = %q, want P=867_0002", requestedQuery)
	}
	if page.ID != "867_0002" {
		t.Errorf("page ID = %q, want 867_0002", page.ID)
	}
	if !bytes.Equal(page.ImageData, samplePNG) {
		t.Errorf("image data = %v, want %v", page.ImageData, samplePNG)
	}
	if page.MimeType != "image/png" {
		t.Errorf("mime type = %q, want image/png", page.MimeType)
	}
	if page.FileExtension() != ".png" {
		t.Errorf("extension = %q, want .png", page.FileExtension())
	}
}

func TestFetchTeletextPageRejectsInvalidPageID(t *testing.T) {
	for _, pageID := range []string{"", "867", "867_2", "abc_0002", "867_0002; rm -rf /"} {
		if _, err := fetchTeletextPage("http://example.invalid", pageID); err == nil {
			t.Errorf("expected error for page id %q", pageID)
		}
	}
}

func TestFetchTeletextPageHandlesErrorResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "non-200 http status", status: http.StatusNotFound, body: "not found"},
		{name: "api error code", status: http.StatusOK, body: `{"meta":{"code":"404"},"data":[]}`},
		{name: "empty data", status: http.StatusOK, body: `{"meta":{"code":"200"},"data":[]}`},
		{name: "no image in content", status: http.StatusOK, body: `{"meta":{"code":"200"},"data":[{"content":{"image":"<img src=\"nope\">"}}]}`},
		{name: "malformed json", status: http.StatusOK, body: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			if _, err := fetchTeletextPage(server.URL, "867_0002"); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestTeletextFileExtensionFollowsMimeType(t *testing.T) {
	tests := map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/gif":  ".gif",
		"":           ".png",
	}

	for mimeType, want := range tests {
		page := &TeletextPage{MimeType: mimeType}
		if got := page.FileExtension(); got != want {
			t.Errorf("FileExtension() for %q = %q, want %q", mimeType, got, want)
		}
	}
}
