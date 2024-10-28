package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/google/uuid"
)

type VideoPostprocessingHandler struct {
	next ContextHandler
}

func cutVideo(input, output string, startSeconds, durationSeconds float64) error {
	args := []string{}
	if startSeconds > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.4f", startSeconds))
	} else if startSeconds < 0 {
		args = append(args, "-sseof", fmt.Sprintf("%.4f", startSeconds))
	}
	args = append(args, "-i", input)
	if durationSeconds > 0 {
		args = append(args, "-t", fmt.Sprintf("%.4f", durationSeconds))
	}
	args = append(args, output)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (u *VideoPostprocessingHandler) execute(m *Context) {
	log.Println("Entering VideoPostprocessingHandler")
	shouldTryPostprocessing := <-m.cutVideoArgsParsed
	if m.action == DownloadVideo && shouldTryPostprocessing {
		startSeconds := <-m.startSeconds
		durationSeconds := <-m.durationSeconds
		videoID := uuid.New().String()
		filePath := fmt.Sprintf("/tmp/%s.mp4", videoID)

		err := cutVideo(m.originalVideoPath, filePath, startSeconds, durationSeconds)
		if err != nil {
			log.Println(err)
		} else {
			m.possiblyProcessedVideoPath = filePath
		}
		// Execute ffprobe to get video duration
		// ffprobeCmd := exec.Command("ffprobe",
		// 	"-v", "error",
		// 	"-show_entries", "format=duration",
		// 	"-of", "default=noprint_wrappers=1:nokey=1",
		// 	m.originalVideoPath)

		// var out bytes.Buffer
		// ffprobeCmd.Stdout = &out

		// if err := ffprobeCmd.Run(); err != nil {
		// 	log.Fatalf("Failed to execute ffprobe: %v", err)
		// }

		// // Parse video duration
		// videoDuration, err := strconv.ParseFloat(strings.TrimSpace(out.String()), 64)
		// if err != nil {
		// 	log.Fatalf("Failed to parse video duration: %v", err)
		// }

		// // Prepare ffmpeg command
		// ffmpegCmd := exec.Command("ffmpeg",
		// 	"-ss", fmt.Sprintf("%.2f", sanitizedStartSeconds))

		// if durationSeconds != nil {
		// 	ffmpegCmd.Args = append(ffmpegCmd.Args, "-t", fmt.Sprintf("%.2f", *durationSeconds))
		// }

		// ffmpegCmd.Args = append(ffmpegCmd.Args, "-i", pathIn, pathOut)

		// // Uncomment for debugging
		// // ffmpegCmd.Args = append(ffmpegCmd.Args, "-loglevel", "debug", "-report")

		// output, err := ffmpegCmd.CombinedOutput()
		// if err != nil {
		// 	log.Fatalf("ffmpeg failed: %v", err)
		// }

		// log.Printf("ffmpeg output: %s", string(output))
	}
	u.next.execute(m)
}

func (u *VideoPostprocessingHandler) setNext(next ContextHandler) {
	u.next = next
}

// func main() {
// 	// Assuming `startSeconds`, `durationSeconds`, `pathIn`, `pathOut` are defined.
// 	var startSeconds, durationSeconds *float64 // replace with actual values
// 	pathIn := "input.mp4"                     // replace with actual path
// 	pathOut := "output.mp4"                    // replace with actual path

// 	var sanitizedStartSeconds float64

// 	if startSeconds != nil && *startSeconds >= 0.0 {
// 		sanitizedStartSeconds = *startSeconds
// 	} else {
// 		// Execute ffprobe to get video duration
// 		ffprobeCmd := exec.Command("ffprobe",
// 			"-v", "error",
// 			"-show_entries", "format=duration",
// 			"-of", "default=noprint_wrappers=1:nokey=1",
// 			pathIn)

// 		var out bytes.Buffer
// 		ffprobeCmd.Stdout = &out

// 		if err := ffprobeCmd.Run(); err != nil {
// 			log.Fatalf("Failed to execute ffprobe: %v", err)
// 		}

// 		// Parse video duration
// 		videoDuration, err := strconv.ParseFloat(strings.TrimSpace(out.String()), 64)
// 		if err != nil {
// 			log.Fatalf("Failed to parse video duration: %v", err)
// 		}

// 		sanitizedStartSeconds = videoDuration + *startSeconds
// 	}

// 	// Prepare ffmpeg command
// 	ffmpegCmd := exec.Command("ffmpeg",
// 		"-ss", fmt.Sprintf("%.2f", sanitizedStartSeconds))

// 	if durationSeconds != nil {
// 		ffmpegCmd.Args = append(ffmpegCmd.Args, "-t", fmt.Sprintf("%.2f", *durationSeconds))
// 	}

// 	ffmpegCmd.Args = append(ffmpegCmd.Args, "-i", pathIn, pathOut)

// 	// Uncomment for debugging
// 	// ffmpegCmd.Args = append(ffmpegCmd.Args, "-loglevel", "debug", "-report")

// 	output, err := ffmpegCmd.CombinedOutput()
// 	if err != nil {
// 		log.Fatalf("ffmpeg failed: %v", err)
// 	}

// 	log.Printf("ffmpeg output: %s", string(output))
// }
