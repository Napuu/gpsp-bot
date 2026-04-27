package doctor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/napuu/gpsp-bot/internal/config"
)

// Run executes the doctor command and prints a report.
func Run() {
	// Define symbols and their colors
	const (
		successSymbol = "[✓]"
		errorSymbol   = "[✗]"
		warningSymbol = "[!]"
	)

	// ANSI color codes
	const (
		green  = "\033[32m"
		red    = "\033[31m"
		yellow = "\033[33m"
		reset  = "\033[0m"
	)

	var report []string

	// === Enabled Features ===
	report = append(report, "=== Enabled Features ===")
	report = append(report, checkEnabledFeatures(warningSymbol, errorSymbol, successSymbol))

	// === External Services ===
	report = append(report, "\n=== External Services ===")
	report = append(report, checkExternalServices(warningSymbol, errorSymbol, successSymbol)...)

	// === Extractors ===
	report = append(report, "\n=== Extractors ===")
	report = append(report, checkExtractors(errorSymbol, successSymbol, warningSymbol)...)

	// === Proxies ===
	proxyResults := checkProxies(warningSymbol, errorSymbol, successSymbol)
	if len(proxyResults) > 0 {
		report = append(report, "\n=== Proxies ===")
		report = append(report, proxyResults...)
	}

	// === Directories ===
	report = append(report, "\n=== Directories ===")
	report = append(report, checkDirectories(errorSymbol, successSymbol)...)

	// Print report with colored symbols and aligned comments
	maxWidth := 0
	ansiEscapeRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for _, line := range report {
		// Strip ANSI color codes for width calculation
		strippedLine := ansiEscapeRegex.ReplaceAllString(line, "")
		commentPos := strings.Index(strippedLine, "#")
		if commentPos == -1 {
			continue
		}
		if commentPos > maxWidth {
			maxWidth = commentPos
		}
	}

	for _, line := range report {
		strippedLine := ansiEscapeRegex.ReplaceAllString(line, "")
		commentPos := strings.Index(strippedLine, "#")
		if commentPos == -1 {
			switch {
			case strings.HasPrefix(line, successSymbol):
				fmt.Printf("%s%s%s%s\n", green, successSymbol, reset, line[len(successSymbol):])
			case strings.HasPrefix(line, errorSymbol):
				fmt.Printf("%s%s%s%s\n", red, errorSymbol, reset, line[len(errorSymbol):])
			case strings.HasPrefix(line, warningSymbol):
				fmt.Printf("%s%s%s%s\n", yellow, warningSymbol, reset, line[len(warningSymbol):])
			default:
				fmt.Println(line)
			}
			continue
		}

		// Pad the line to align comments
		nonCommentPart := strippedLine[:commentPos]
		commentPart := strippedLine[commentPos:]
		paddedLine := fmt.Sprintf("%-*s%s", maxWidth, nonCommentPart, commentPart)

		// Reapply ANSI colors
		if strings.HasPrefix(line, successSymbol) {
			fmt.Printf("%s%s%s%s\n", green, successSymbol, reset, paddedLine[len(successSymbol):])
		} else if strings.HasPrefix(line, errorSymbol) {
			fmt.Printf("%s%s%s%s\n", red, errorSymbol, reset, paddedLine[len(errorSymbol):])
		} else if strings.HasPrefix(line, warningSymbol) {
			fmt.Printf("%s%s%s%s\n", yellow, warningSymbol, reset, paddedLine[len(warningSymbol):])
		} else {
			fmt.Println(paddedLine)
		}
	}
}

// checkTelegramToken validates the Telegram token.
func checkTelegramToken(errorSymbol, successSymbol string) string {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return fmt.Sprintf("%s Telegram Token (TELEGRAM_TOKEN): Missing # Handles Telegram bot interactions", errorSymbol)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s Telegram Token (TELEGRAM_TOKEN): Invalid # Handles Telegram bot interactions", errorSymbol)
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.OK {
		return fmt.Sprintf("%s Telegram Token (TELEGRAM_TOKEN): Invalid # Handles Telegram bot interactions", errorSymbol)
	}

	sanitizedToken := "****" + token[len(token)-4:]
return fmt.Sprintf("%s Telegram Token (TELEGRAM_TOKEN): Valid (Bot: @%s, Token: %s) # Handles Telegram bot interactions",
	successSymbol, result.Result.Username, sanitizedToken)
}

// checkDiscordToken validates the Discord token.
func checkDiscordToken(errorSymbol, successSymbol string) string {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Sprintf("%s Discord Token (DISCORD_TOKEN): Missing # Handles Discord bot interactions", errorSymbol)
	}

	url := "https://discord.com/api/v10/users/@me"
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bot "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s Discord Token (DISCORD_TOKEN): Invalid # Handles Discord bot interactions", errorSymbol)
	}

	var result struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("%s Discord Token (DISCORD_TOKEN): Invalid # Handles Discord bot interactions", errorSymbol)
	}

	sanitizedToken := "****" + token[len(token)-4:]
return fmt.Sprintf("%s Discord Token (DISCORD_TOKEN): Valid (Bot: %s, Token: %s) # Handles Discord bot interactions",
	successSymbol, result.Username, sanitizedToken)
}

// checkMistralToken validates the Mistral token.
func checkMistralToken(warningSymbol, errorSymbol, successSymbol string) string {
	token := os.Getenv("MISTRAL_TOKEN")
if token == "" {
		return fmt.Sprintf("%s Mistral Token (MISTRAL_TOKEN): Missing # Required for tuplilla (LLM) feature", warningSymbol)
	}

	url := "https://api.mistral.ai/v1/models"
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s Mistral Token (MISTRAL_TOKEN): Invalid # Required for tuplilla (LLM) feature", errorSymbol)
	}

	sanitizedToken := "****" + token[len(token)-4:]
return fmt.Sprintf("%s Mistral Token (MISTRAL_TOKEN): Valid (Token: %s) # Required for tuplilla (LLM) feature",
	successSymbol, sanitizedToken)
}

// checkYtDlp checks if yt-dlp is installed and available.
func checkYtDlp(errorSymbol, successSymbol string) string {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Sprintf("%s yt-dlp: Not installed", errorSymbol)
	}

	version, err := exec.Command("yt-dlp", "--version").Output()
	if err != nil {
		return fmt.Sprintf("%s yt-dlp: Installed but not executable", errorSymbol)
	}

	return fmt.Sprintf("%s yt-dlp: Installed (v%s) # Generic video/audio extractor (supports YouTube, Vimeo, etc.)", successSymbol, strings.TrimSpace(string(version)))
}

// checkYleDl checks if yle-dl is installed and available.
func checkYleDl(errorSymbol, successSymbol, warningSymbol string) string {
	_, err := exec.LookPath("yle-dl")
	if err != nil {
		return fmt.Sprintf("%s yle-dl: Not installed # Downloads videos from YLE (optional)", warningSymbol)
	}

	version, err := exec.Command("yle-dl", "--version").Output()
	if err != nil {
		return fmt.Sprintf("%s yle-dl: Installed but not executable # Downloads videos from YLE (optional)", warningSymbol)
	}

	return fmt.Sprintf("%s yle-dl: Installed (v%s) # Downloads videos from YLE (optional)", successSymbol, strings.TrimSpace(string(version)))
}

// checkFfmpeg checks if ffmpeg is installed and available.
func checkFfmpeg(errorSymbol, successSymbol string) string {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Sprintf("%s ffmpeg: Not installed", errorSymbol)
	}

	version, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		return fmt.Sprintf("%s ffmpeg: Installed but not executable", errorSymbol)
	}

	versionLine := strings.Split(string(version), "\n")[0]
	versionParts := strings.Fields(versionLine)
	if len(versionParts) >= 3 {
		return fmt.Sprintf("%s ffmpeg: Installed (v%s) # Processes audio/video files", successSymbol, versionParts[2])
	}

	return fmt.Sprintf("%s ffmpeg: Installed # Processes audio/video files", successSymbol)
}

// checkProxies validates SOCKS proxies from PROXY_URLS.
func checkProxies(warningSymbol, errorSymbol, successSymbol string) []string {
	proxyURLs := config.ProxyUrls()
	if len(proxyURLs) == 0 || (len(proxyURLs) == 1 && proxyURLs[0] == "") {
		return []string{fmt.Sprintf("%s SOCKS Proxies (PROXY_URLS): None configured", warningSymbol)}
	}

	var results []string
	for _, proxy := range proxyURLs {
		proxy = strings.TrimSpace(proxy)
		if proxy == "" {
			continue
		}

		cmd := exec.Command("curl", "--max-time", "5", "--socks5", proxy, "ip.me")
		output, err := cmd.Output()
		if err == nil {
			exitIP := strings.TrimSpace(string(output))
			results = append(results, fmt.Sprintf("%s SOCKS Proxy: %s (Exit IP: %s)", successSymbol, proxy, exitIP))
		} else {
			results = append(results, fmt.Sprintf("%s SOCKS Proxy: %s (Failed)", errorSymbol, proxy))
		}
	}

	return results
}

// checkDirectory checks if a directory exists and is writable.
func checkDirectory(name, defaultPath, errorSymbol, successSymbol string) string {
	path := os.Getenv(name)
	if path == "" {
		path = defaultPath
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Sprintf("%s %s: Invalid path (%s)", errorSymbol, name, path)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Try to create the directory
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Sprintf("%s %s: Directory does not exist and cannot be created (%s)", errorSymbol, name, absPath)
		}
	}

	// Check if directory is writable
	testFile := filepath.Join(absPath, ".doctor_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Sprintf("%s %s: Not writable (%s)", errorSymbol, name, absPath)
	}

	// Clean up
	os.Remove(testFile)

	defaultMarker := ""
	if path == defaultPath {
		defaultMarker = " (Default)"
	}

	explanation := ""
	switch name {
	case "YTDLP_TMP_DIR":
		explanation = "# Temp files for yt-dlp"
	case "EURIBOR_GRAPH_DIR":
		explanation = "# Stores Euribor graphs"
	case "EURIBOR_CSV_DIR":
		explanation = "# Stores Euribor CSV exports"
	case "REPOST_DB_DIR":
		explanation = "# Stores repost detection database"
	}

	return fmt.Sprintf("%s %s: Writable (%s)%s %s", successSymbol, name, absPath, defaultMarker, explanation)
}

// checkEnabledFeatures validates ENABLED_FEATURES.
func checkEnabledFeatures(warningSymbol, errorSymbol, successSymbol string) string {
	enabledFeatures := config.EnabledFeatures()
	if len(enabledFeatures) == 0 || (len(enabledFeatures) == 1 && enabledFeatures[0] == "") {
		return fmt.Sprintf("%s ENABLED_FEATURES: Not set # Valid features: ping, dl, euribor, tuplilla, stats, version", errorSymbol)
	}

	validFeatures := map[string]bool{
		"ping":     true,
		"dl":       true,
		"euribor":  true,
		"tuplilla": true,
		"stats":    true,
		"version":  true,
	}

	invalidFeatures := []string{}
	for _, feature := range enabledFeatures {
		if !validFeatures[feature] {
			invalidFeatures = append(invalidFeatures, feature)
		}
	}

	if len(invalidFeatures) > 0 {
		return fmt.Sprintf("%s ENABLED_FEATURES: Unknown feature(s) %q # Valid features: ping, dl, euribor, tuplilla, stats, version", warningSymbol, strings.Join(invalidFeatures, ", "))
	}

	return fmt.Sprintf("%s ENABLED_FEATURES: Valid (%s)", successSymbol, strings.Join(enabledFeatures, ", "))
}

// checkExternalServices validates external service tokens.
func checkExternalServices(warningSymbol, errorSymbol, successSymbol string) []string {
	var results []string
	results = append(results, checkTelegramToken(errorSymbol, successSymbol))
	results = append(results, checkDiscordToken(errorSymbol, successSymbol))
	results = append(results, checkMistralToken(warningSymbol, errorSymbol, successSymbol))
	return results
}

// checkExtractors validates extractor tools.
func checkExtractors(errorSymbol, successSymbol, warningSymbol string) []string {
	var results []string
	results = append(results, checkYtDlp(errorSymbol, successSymbol))
	results = append(results, checkFfmpeg(errorSymbol, successSymbol))
	results = append(results, checkYleDl(errorSymbol, successSymbol, warningSymbol))
	return results
}

// checkDirectories validates directory permissions.
func checkDirectories(errorSymbol, successSymbol string) []string {
	var results []string
	results = append(results, checkDirectory("YTDLP_TMP_DIR", "/tmp/ytdlp", errorSymbol, successSymbol))
	results = append(results, checkDirectory("EURIBOR_GRAPH_DIR", "/tmp/euribor-graphs", errorSymbol, successSymbol))
	results = append(results, checkDirectory("EURIBOR_CSV_DIR", "/tmp/euribor-exports", errorSymbol, successSymbol))
	results = append(results, checkDirectory("REPOST_DB_DIR", "/tmp/repost-db", errorSymbol, successSymbol))
	return results
}
