package utils

import (
	"fmt"
	"log/slog"
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

func cycleProxy() string {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	proxy := proxyURLs[currentProxy]
	currentProxy = (currentProxy + 1) % len(proxyURLs)
	return proxy
}

// isYleURL checks if the URL is from yle.fi domain
func isYleURL(url string) bool {
	return strings.Contains(url, "yle.fi")
}

// isYleDlAvailable checks if yle-dl is available in PATH
func isYleDlAvailable() bool {
	_, err := exec.LookPath("yle-dl")
	return err == nil
}

// attemptYleDlDownload tries to download using yle-dl
func attemptYleDlDownload(url, filePath, proxy string) bool {
	args := []string{
		"--output", filePath,
	}

	if proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	args = append(args, url)

	if proxy != "" {
		slog.Info(fmt.Sprintf("Attempting download with yle-dl using proxy %s", proxy))
	} else {
		slog.Info("Attempting download with yle-dl")
	}
	cmd := exec.Command("yle-dl", args...)
	err := cmd.Run()
	if err != nil {
		slog.Info(fmt.Sprintf("yle-dl download failed: %v", err))
		return false
	}
	return true
}

func DownloadVideo(url string, targetSizeInMB uint64) string {
	tmpPath := config.FromEnv().YTDLP_TMP_DIR
	videoID := uuid.New().String()
	filePath := fmt.Sprintf("%s/%s.mp4", tmpPath, videoID)

	return tryDownloadWithProxies(url, filePath, targetSizeInMB)
}

// tryDownloadWithProxies attempts download with yle-dl (if yle.fi link) and yt-dlp, cycling through proxies
func tryDownloadWithProxies(url, filePath string, targetSizeInMB uint64) string {
	// For yle.fi URLs, try yle-dl first if available
	if isYleURL(url) && isYleDlAvailable() {
		slog.Info("Attempting download with yle-dl (no proxy)")
		if attemptYleDlDownload(url, filePath, "") {
			return filePath
		}

		// Try yle-dl with proxies
		for i := 0; i < len(proxyURLs); i++ {
			proxy := cycleProxy()
			slog.Info(fmt.Sprintf("yle-dl download failed, trying with proxy %s", proxy))

			if attemptYleDlDownload(url, filePath, proxy) {
				return filePath
			}
		}

		slog.Info("yle-dl failed with all proxies, falling back to yt-dlp")
	}

	// Try yt-dlp (for all URLs, or as fallback for yle.fi)
	slog.Info("Downloading with yt-dlp (no proxy)")
	if attemptYtDlpDownload(url, filePath, "", targetSizeInMB) {
		return filePath
	}

	// Try yt-dlp with proxies
	for i := 0; i < len(proxyURLs); i++ {
		proxy := cycleProxy()
		slog.Info(fmt.Sprintf("yt-dlp download failed, trying with proxy %s", proxy))

		if attemptYtDlpDownload(url, filePath, proxy, targetSizeInMB) {
			return filePath
		}
	}

	slog.Info("Download failed with all methods")
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
