package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

func downloadVideo(url string, targetSizeInMB uint64) string {
	videoID := uuid.New().String()
	filePath := fmt.Sprintf("/tmp/%s.mp4", videoID)
	attemptDownload(url, filePath, "", targetSizeInMB)
	// proxyURLs := strings.Split(getConfigValue(EnvVariableSocksURLs), ";")

	// for _, proxy := range proxyURLs {
	// 	targetID := simpleXOR(url)
	// 	log.Printf("yt-dlp trying with proxy %s, target %x", proxy, targetID)

	// 	if attemptDownload(url, filePath, proxy, targetSizeInMB) {
	// 		log.Printf("yt-dlp success with proxy %s, target %x", proxy, targetID)
	// 		outputPath, err := filepath.Abs(filePath)
	// 		if err != nil {
	// 			log.Printf("Error resolving file path: %v", err)
	// 			return nil
	// 		}
	// 		return &outputPath
	// 	}
	// }

	return filePath
}

func attemptDownload(url, filePath, proxy string, targetSizeInMB uint64) bool {
	cmd := exec.Command("yt-dlp",
	// no proxy for now
	// "--proxy", proxy,
	"-f", fmt.Sprintf("((bv*[filesize<=%dM]/bv*)[height<=720]/(wv*[filesize<=%dM]/wv*)) + ba / (b[filesize<=%dM]/b)[height<=720]/(w[filesize<=%dM]/w)",
	targetSizeInMB, targetSizeInMB, targetSizeInMB, targetSizeInMB),
	// "--limit-rate", "50K",
	"-S", "codec:h264",
	"--merge-output-format", "mp4",
	"--recode", "mp4",
	"-o", filePath,
	url,
)

output, err := cmd.CombinedOutput()
log.Printf("yt-dlp output:\n%s", string(output))

if err != nil {
	log.Printf("yt-dlp failed: %v\nstdout+stderr: %s", err, string(output))
	return false
}
return true
}

// Placeholder functions for configuration and ID generation.
func getConfigValue(key string) string {
	return os.Getenv(key)
}

func simpleXOR(s string) int {
	xor := 0
	for _, c := range s {
		xor ^= int(c)
	}
	return xor
}

const EnvVariableSocksURLs = "SOCKS_URLS"

// func main() {
// 	// Example usage:
// 	url := "https://example.com/video"
// 	targetSizeInMB := uint64(50)

// 	result := downloadVideo(url, targetSizeInMB)
// 	if result != nil {
// 		fmt.Printf("Download succeeded, file path: %s\n", *result)
// 	} else {
// 		fmt.Println("Download failed.")
// 	}
// }

