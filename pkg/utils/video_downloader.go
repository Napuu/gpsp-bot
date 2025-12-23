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
func attemptYleDlDownload(url, filePath string) bool {
	args := []string{
		"--output", filePath,
		url,
	}

	slog.Info("Attempting download with yle-dl")
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

	// Try yle-dl first for yle.fi URLs if available
	if isYleURL(url) && isYleDlAvailable() {
		if attemptYleDlDownload(url, filePath) {
			return filePath
		}
		slog.Info("yle-dl failed, falling back to yt-dlp")
	}

	slog.Info("Downloading with no proxy")
	if attemptDownload(url, filePath, "", targetSizeInMB) {
		return filePath
	}

	for i := 0; i < len(proxyURLs); i++ {
		proxy := cycleProxy()
		slog.Info(fmt.Sprintf("Downloading with no proxy failed, trying with %s", proxy))

		if attemptDownload(url, filePath, proxy, targetSizeInMB) {
			return filePath
		}
	}

	slog.Info("Downloading failed")
	return ""
}

func attemptDownload(url, filePath, proxy string, targetSizeInMB uint64) bool {
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
