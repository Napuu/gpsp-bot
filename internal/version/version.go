package version

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Version is set at build time via ldflags
// Format: yyyymmddhhmmss (e.g., 20231128151500) or semantic version (e.g., v1.0.0)
var Version = "dev"

func getYtDlpVersion() string {
	out, err := exec.Command("yt-dlp", "--version").Output()
	if err != nil {
		return "not found"
	}
	return strings.TrimSpace(string(out))
}

func getFfmpegVersion() string {
	out, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		return "not found"
	}
	// First line is like "ffmpeg version 6.0 Copyright ..."
	firstLine := strings.SplitN(string(out), "\n", 2)[0]
	const prefix = "ffmpeg version "
	if idx := strings.Index(firstLine, prefix); idx >= 0 {
		rest := firstLine[idx+len(prefix):]
		if parts := strings.Fields(rest); len(parts) > 0 {
			return parts[0]
		}
	}
	return strings.TrimSpace(firstLine)
}

// GetHumanReadableVersion returns a human-readable version string
// that includes the bot name and, for timestamp versions, a formatted date.
// The actual version tag is always kept visible.
// yt-dlp and ffmpeg versions are appended on separate lines.
func GetHumanReadableVersion() string {
	var botLine string
	if Version == "dev" {
		botLine = "gpsp-bot dev (development build)"
	} else if len(Version) == 14 {
		// Try to parse as timestamp format (yyyymmddhhmmss)
		if t, err := time.Parse("20060102150405", Version); err == nil {
			botLine = fmt.Sprintf("gpsp-bot %s (built %s)", Version, t.UTC().Format("2006-01-02 15:04:05 UTC"))
		} else {
			botLine = fmt.Sprintf("gpsp-bot %s", Version)
		}
	} else {
		// For semantic versions or other formats, just display nicely
		botLine = fmt.Sprintf("gpsp-bot %s", Version)
	}

	return fmt.Sprintf("%s\nyt-dlp: %s\nffmpeg: %s", botLine, getYtDlpVersion(), getFfmpegVersion())
}
