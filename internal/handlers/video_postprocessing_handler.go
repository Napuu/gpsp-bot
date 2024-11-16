package handlers

import (
	"fmt"
	"log"
	"log/slog"
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

func (u *VideoPostprocessingHandler) Execute(m *Context) {
	slog.Debug("Entering VideoPostprocessingHandler")
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
	}
	u.next.Execute(m)
}

func (u *VideoPostprocessingHandler) SetNext(next ContextHandler) {
	u.next = next
}
