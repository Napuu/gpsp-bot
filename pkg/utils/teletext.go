package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const teletextAPIURL = "https://yle.fi/aihe/yle-ttv/json"

// Page identifiers follow YLE's "<page>_<subpage>" format, e.g. "867_0002".
var teletextPageIDPattern = regexp.MustCompile(`^[0-9]{3}_[0-9]{4}$`)

// The API returns the rendered page wrapped in an HTML img tag with a data URI payload.
var teletextDataURIPattern = regexp.MustCompile(`data:image/(png|jpeg|gif);base64,([A-Za-z0-9+/=\s]+)`)

var whitespacePattern = regexp.MustCompile(`\s`)

type teletextResponse struct {
	Meta struct {
		Code string `json:"code"`
	} `json:"meta"`
	Data []struct {
		Content struct {
			Image string `json:"image"`
		} `json:"content"`
	} `json:"data"`
}

type TeletextPage struct {
	ID        string
	ImageData []byte
	MimeType  string
}

// FetchTeletextPage fetches a single YLE teletext page as a decoded image.
func FetchTeletextPage(pageID string) (*TeletextPage, error) {
	return fetchTeletextPage(teletextAPIURL, pageID)
}

func IsValidTeletextPageID(pageID string) bool {
	return teletextPageIDPattern.MatchString(pageID)
}

func fetchTeletextPage(baseURL, pageID string) (*TeletextPage, error) {
	if !IsValidTeletextPageID(pageID) {
		return nil, fmt.Errorf("invalid teletext page id %q, expected format like 867_0002", pageID)
	}

	requestURL := baseURL + "?P=" + url.QueryEscape(pageID)
	slog.Debug("Fetching teletext page", "url", requestURL)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("teletext request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("teletext request returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read teletext response: %w", err)
	}

	return parseTeletextPage(pageID, body)
}

func parseTeletextPage(pageID string, body []byte) (*TeletextPage, error) {
	var parsed teletextResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse teletext response: %w", err)
	}

	if parsed.Meta.Code != "" && parsed.Meta.Code != "200" {
		return nil, fmt.Errorf("teletext page %s unavailable, api returned code %s", pageID, parsed.Meta.Code)
	}

	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("teletext page %s contained no data", pageID)
	}

	match := teletextDataURIPattern.FindStringSubmatch(parsed.Data[0].Content.Image)
	if match == nil {
		return nil, fmt.Errorf("teletext page %s contained no image", pageID)
	}

	imageData, err := base64.StdEncoding.DecodeString(whitespacePattern.ReplaceAllString(match[2], ""))
	if err != nil {
		return nil, fmt.Errorf("failed to decode teletext image: %w", err)
	}

	return &TeletextPage{
		ID:        pageID,
		ImageData: imageData,
		MimeType:  "image/" + match[1],
	}, nil
}

// FileExtension returns the extension matching the fetched image, defaulting to png.
func (p *TeletextPage) FileExtension() string {
	switch p.MimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}
