package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/napuu/gpsp-bot/assets"
	"github.com/napuu/gpsp-bot/internal/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	dayOverlayFontSize       = 62
	dayOverlayLineGap        = 12
	dayOverlayPadding        = 24
	dayVideoContentTopRatio  = 0.119 // bottom edge of top black bar in day_template.mp4
	dayOverlayBottomMarginPx = 4
	dayOverlayYOffsetPx      = 250
	dayTemplateAssetName     = "day_template.mp4"
)

// DayVideoTmpDir returns the directory for day meme temp files.
func DayVideoTmpDir() string {
	return filepath.Join(config.FromEnv().YTDLP_TMP_DIR, "day-video")
}

func ordinalSuffix(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// FormatDayOverlayText returns the two overlay lines for the given time.
func FormatDayOverlayText(t time.Time) (dateLine, weekdayLine string) {
	day := t.Day()
	dateLine = fmt.Sprintf("%d%s %s, %d", day, ordinalSuffix(day), t.Format("January"), t.Year())
	weekdayLine = t.Format("Monday")
	return dateLine, weekdayLine
}

func loadDayOverlayFace() (font.Face, error) {
	fontBytes, err := assets.TemplateFS.ReadFile("Roboto-Bold.ttf")
	if err != nil {
		return nil, fmt.Errorf("read overlay font: %w", err)
	}
	fontData, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    dayOverlayFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}
	return face, nil
}

func textLineBounds(face font.Face, text string) (width, height int) {
	advance := font.MeasureString(face, text)
	return advance.Ceil(), face.Metrics().Height.Ceil()
}

// GenerateDayTextPNG renders the date and weekday on a tight white background.
func GenerateDayTextPNG(dateLine, weekdayLine, outputPath string) error {
	face, err := loadDayOverlayFace()
	if err != nil {
		return err
	}
	defer face.Close()

	dateWidth, lineHeight := textLineBounds(face, dateLine)
	weekdayWidth, _ := textLineBounds(face, weekdayLine)
	width := max(dateWidth, weekdayWidth) + 2*dayOverlayPadding
	height := 2*lineHeight + dayOverlayLineGap + 2*dayOverlayPadding

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}

	centerX := func(text string) fixed.Int26_6 {
		textWidth := font.MeasureString(face, text)
		return fixed.I((width - textWidth.Ceil()) / 2)
	}

	drawer.Dot = fixed.Point26_6{
		X: centerX(dateLine),
		Y: fixed.I(dayOverlayPadding + face.Metrics().Ascent.Ceil()),
	}
	drawer.DrawString(dateLine)

	drawer.Dot = fixed.Point26_6{
		X: centerX(weekdayLine),
		Y: fixed.I(dayOverlayPadding + lineHeight + dayOverlayLineGap + face.Metrics().Ascent.Ceil()),
	}
	drawer.DrawString(weekdayLine)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create png: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, rgba); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

func writeEmbeddedAsset(name, dest string) error {
	data, err := assets.TemplateFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read embedded asset %q: %w", name, err)
	}
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("write asset %q: %w", name, err)
	}
	return nil
}

func overlayDayText(templatePath, textPath, outputPath string) error {
	filter := fmt.Sprintf(
		"overlay=x=(W-w)/2:y=H*%.3f-h-%d+%d",
		dayVideoContentTopRatio,
		dayOverlayBottomMarginPx,
		dayOverlayYOffsetPx,
	)
	args := []string{
		"-y",
		"-i", templatePath,
		"-i", textPath,
		"-filter_complex", filter,
		"-c:a", "copy",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("ffmpeg overlay: %w: %s", err, msg)
		}
		return fmt.Errorf("ffmpeg overlay: %w", err)
	}
	return nil
}

// GenerateDayVideo creates a day meme video for the given time.
func GenerateDayVideo(outputPath string, t time.Time) error {
	tmpDir := filepath.Join(DayVideoTmpDir(), uuid.New().String())
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	templatePath := filepath.Join(tmpDir, dayTemplateAssetName)
	if err := writeEmbeddedAsset(dayTemplateAssetName, templatePath); err != nil {
		return err
	}

	dateLine, weekdayLine := FormatDayOverlayText(t)
	textPath := filepath.Join(tmpDir, "text.png")
	if err := GenerateDayTextPNG(dateLine, weekdayLine, textPath); err != nil {
		return err
	}

	if err := overlayDayText(templatePath, textPath, outputPath); err != nil {
		return err
	}
	return nil
}
