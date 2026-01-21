package utils

import (
	"bytes"
	"image"
	"image/draw"
	"os/exec"

	"github.com/disintegration/imaging"
	"github.com/napuu/gpsp-bot/assets"
)

const (
	rectX      = 75
	rectY      = 135
	rectWidth  = 95
	rectHeight = 90
)

// ExtractFirstFrame extracts the first frame from a video using ffmpeg
func ExtractFirstFrame(videoPath, outputPath string) error {
	args := []string{
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// GetTemplateImage reads and decodes the embedded template image
func GetTemplateImage() (image.Image, error) {
	data, err := assets.TemplateFS.ReadFile("pate_template.jpg")
	if err != nil {
		return nil, err
	}

	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// CompositeRepostImage composites the frame into the template at the specified coordinates
func CompositeRepostImage(templateImg image.Image, framePath, outputPath string) error {
	// Load the frame image
	frameImg, err := imaging.Open(framePath)
	if err != nil {
		return err
	}

	// Resize frame to fit the rectangle
	resizedFrame := imaging.Resize(frameImg, rectWidth, rectHeight, imaging.Lanczos)

	// Create a mutable copy of the template
	result := imaging.Clone(templateImg)

	// Draw the resized frame onto the result
	targetRect := image.Rect(rectX, rectY, rectX+rectWidth, rectY+rectHeight)
	draw.Draw(result, targetRect, resizedFrame, image.Point{0, 0}, draw.Over)

	// Save the final image
	err = imaging.Save(result, outputPath)
	if err != nil {
		return err
	}

	return nil
}
