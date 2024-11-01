package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/uuid"
)


var (
	proxyURLs     = strings.Split(os.Getenv("PROXY_URLS"), ";")
	currentProxy  int
	proxyMutex    sync.Mutex
)

func simpleXOR(s string) int {
	xor := 0
	for _, c := range s {
		xor ^= int(c)
	}
	return xor
}

func cycleProxy() string {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	proxy := proxyURLs[currentProxy]
	currentProxy = (currentProxy + 1) % len(proxyURLs)
	return proxy
}

func DownloadVideo(url string, targetSizeInMB uint64) string {
	videoID := uuid.New().String()
	filePath := fmt.Sprintf("/tmp/%s.mp4", videoID)

	if attemptDownload(url, filePath, "", targetSizeInMB) {
		return filePath
	}

	for i := 0; i < len(proxyURLs); i++ {
		proxy := cycleProxy()
		targetID := simpleXOR(url)
		log.Printf("yt-dlp trying with proxy \"%s\", target %x", proxy, targetID)

		if attemptDownload(url, filePath, proxy, targetSizeInMB) {
			log.Printf("yt-dlp success with proxy %s, target %x", proxy, targetID)
			return filePath
		}
	}

	log.Printf("Failed to download video from URL %s after trying all proxies", url)
	return ""
}

func attemptDownload(url, filePath, proxy string, targetSizeInMB uint64) bool {
	args := []string{
		"-f", fmt.Sprintf("((bv*[filesize<=%dM]/bv*)[height<=720]/(wv*[filesize<=%dM]/wv*)) + ba / (b[filesize<=%dM]/b)[height<=720]/(w[filesize<=%dM]/w)",
			targetSizeInMB, targetSizeInMB, targetSizeInMB, targetSizeInMB),
		"-S", "codec:h264",
		"--merge-output-format", "mp4",
		"--recode", "mp4",
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
