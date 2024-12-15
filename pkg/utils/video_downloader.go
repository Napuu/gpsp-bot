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
	proxyURLs    = strings.Split(config.FromEnv().PROXY_URLS, ";")
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

func DownloadVideo(url string, targetSizeInMB uint64) string {
	tmpPath := config.FromEnv().YTDLP_TMP_DIR
	videoID := uuid.New().String()
	filePath := fmt.Sprintf("%s/%s.mp4", tmpPath, videoID)

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
