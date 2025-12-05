package version

import (
	"fmt"
	"time"
)

// Version is set at build time via ldflags
// Format: yyyymmddhhmmss (e.g., 20231128151500) or semantic version (e.g., v1.0.0)
var Version = "dev"

// GetHumanReadableVersion returns a human-readable version string
// that includes the bot name and, for timestamp versions, a formatted date.
// The actual version tag is always kept visible.
func GetHumanReadableVersion() string {
	if Version == "dev" {
		return "gpsp-bot dev (development build)"
	}

	// Try to parse as timestamp format (yyyymmddhhmmss)
	if len(Version) == 14 {
		if t, err := time.Parse("20060102150405", Version); err == nil {
			return fmt.Sprintf("gpsp-bot %s (built %s)", Version, t.UTC().Format("2006-01-02 15:04:05 UTC"))
		}
	}

	// For semantic versions or other formats, just display nicely
	return fmt.Sprintf("gpsp-bot %s", Version)
}
