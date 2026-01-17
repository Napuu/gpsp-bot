package utils

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/internal/config"
)

var (
	proxyURLs    = config.ProxyUrls()
	currentProxy int
	proxyMutex   sync.Mutex
)

const (
	// maxFileSizeBytes is the maximum file size for HTTP downloads (500 MB)
	// This matches yt-dlp's --max-filesize behavior
	maxFileSizeBytes = 500 * 1024 * 1024
)

// Common user agents to avoid being blocked
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
}

type ExtractorFunc func(url, filePath, proxy string, targetSizeInMB uint64) bool

type SpecialExtractor struct {
	Command       string
	URLMatcher    func(string) bool
	DownloadFunc  ExtractorFunc
	SupportsProxy bool
}

func isYleFiURL(urlStr string) bool {
	parsedURL, err := neturl.Parse(urlStr)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsedURL.Hostname())
	return host == "yle.fi" || strings.HasSuffix(host, ".yle.fi")
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func attemptYleDlDownload(urlStr, filePath, proxy string, targetSizeInMB uint64) bool {
	args := []string{
		"-o", filePath,
		urlStr,
	}

	cmd := exec.Command("yle-dl", args...)
	err := cmd.Run()
	return err == nil
}

func getSpecialExtractor(urlStr string) SpecialExtractor {
	extractors := []SpecialExtractor{
		{
			Command:       "yle-dl",
			URLMatcher:    isYleFiURL,
			DownloadFunc:  attemptYleDlDownload,
			SupportsProxy: false,
		},
	}

	for _, extractor := range extractors {
		if extractor.URLMatcher(urlStr) && isCommandAvailable(extractor.Command) {
			return extractor
		}
	}

	return SpecialExtractor{}
}

func cycleProxy() string {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	proxy := proxyURLs[currentProxy]
	currentProxy = (currentProxy + 1) % len(proxyURLs)
	return proxy
}

func tryDownloadWithExtractor(extractor ExtractorFunc, urlStr, filePath string, targetSizeInMB uint64, supportsProxy bool) bool {
	slog.Info("Downloading with no proxy")
	if extractor(urlStr, filePath, "", targetSizeInMB) {
		return true
	}

	if !supportsProxy {
		return false
	}

	for i := 0; i < len(proxyURLs); i++ {
		proxy := cycleProxy()
		slog.Info(fmt.Sprintf("Trying with proxy %s", proxy))

		if extractor(urlStr, filePath, proxy, targetSizeInMB) {
			return true
		}
	}

	return false
}

func DownloadVideo(url string, targetSizeInMB uint64) string {
	tmpPath := config.FromEnv().YTDLP_TMP_DIR
	videoID := uuid.New().String()
	filePath := fmt.Sprintf("%s/%s.mp4", tmpPath, videoID)

	specialExtractor := getSpecialExtractor(url)
	if specialExtractor.Command != "" {
		slog.Info(fmt.Sprintf("Using special extractor: %s", specialExtractor.Command))
		if tryDownloadWithExtractor(specialExtractor.DownloadFunc, url, filePath, targetSizeInMB, specialExtractor.SupportsProxy) {
			return filePath
		}
		slog.Info(fmt.Sprintf("%s failed, falling back to yt-dlp", specialExtractor.Command))
	}

	slog.Info("Using yt-dlp")
	if tryDownloadWithExtractor(attemptYtDlpDownload, url, filePath, targetSizeInMB, true) {
		return filePath
	}

	slog.Info("yt-dlp failed, trying HTTP fallback")
	if tryDownloadWithExtractor(attemptHTTPDownload, url, filePath, targetSizeInMB, true) {
		return filePath
	}

	slog.Info("Downloading failed")
	return ""
}

func attemptYtDlpDownload(url, filePath, proxy string, targetSizeInMB uint64) bool {
	args := []string{
		"-f", fmt.Sprintf("((bv*[filesize<=%dM]/bv*)[height<=720]/(wv*[filesize<=%dM]/wv*)) + ba / (b[filesize<=%dM]/b)[height<=720]/(w[filesize<=%dM]/w)",
			targetSizeInMB, targetSizeInMB, targetSizeInMB, targetSizeInMB),
		"-S", "codec:h264",
		"--merge-output-format", "mp4",
		"--recode", "mp4",
		"--max-filesize", "500M", // just in case
		"-o", filePath,
		url,
	}

	if proxy != "" {
		args = append([]string{"--proxy", proxy}, args...)
	}

	cmd := exec.Command("yt-dlp", args...)
	err := cmd.Run()
	return err == nil
}

func attemptHTTPDownload(url, filePath, proxy string, targetSizeInMB uint64) bool {
	client := &http.Client{}

	if proxy != "" {
		proxyURL, err := neturl.Parse(proxy)
		if err != nil {
			slog.Info(fmt.Sprintf("Failed to parse proxy URL: %v", err))
			return false
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Info(fmt.Sprintf("Failed to create request: %v", err))
		return false
	}

	// Pick a random user agent
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		slog.Info(fmt.Sprintf("HTTP GET failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Info(fmt.Sprintf("HTTP GET returned status: %d", resp.StatusCode))
		return false
	}

	out, err := os.Create(filePath)
	if err != nil {
		slog.Info(fmt.Sprintf("Failed to create file: %v", err))
		return false
	}
	defer out.Close()

	// Download up to 500 MB (matching yt-dlp's --max-filesize behavior)
	// If file is larger, truncate it at the limit
	limitedReader := io.LimitReader(resp.Body, maxFileSizeBytes)

	if resp.ContentLength > 0 && resp.ContentLength > maxFileSizeBytes {
		slog.Info(fmt.Sprintf("File size %d bytes exceeds limit, downloading first 500 MB", resp.ContentLength))
	}

	_, err = io.Copy(out, limitedReader)
	if err != nil {
		slog.Info(fmt.Sprintf("Failed to write file: %v", err))
		// Clean up partially downloaded file on error
		os.Remove(filePath)
		return false
	}

	return true
}
