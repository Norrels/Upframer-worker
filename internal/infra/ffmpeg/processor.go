package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"upframer-worker/internal/domain/entities"
)

type FFmpegProcessor struct{}

func NewFFmpegProcessor() *FFmpegProcessor {
	return &FFmpegProcessor{}
}

func (p *FFmpegProcessor) ProcessVideo(job *entities.VideoJob) (*entities.ProcessingResult, error) {
	outputDir := "frames"

	err := os.MkdirAll(outputDir, 0755)

	if err != nil {
		return &entities.ProcessingResult{
			Success: false,
			Error:   fmt.Errorf("error creating folder: %v", err),
		}, err
	}

	outputName := outputDir + "/frame_%04d.jpg"

	cmd := exec.Command("ffmpeg",
		"-i", job.VideoPath,
		"-vf", "fps=1",
		"-y",
		outputName,
	)

	_, err = cmd.CombinedOutput()

	if err != nil {
		return &entities.ProcessingResult{
			Success: true,
			Error:   fmt.Errorf("error extracting frames: %v", err),
		}, nil
	}

	return &entities.ProcessingResult{
		Success:    true,
		Error:      nil,
		OutputPath: outputDir,
	}, nil
}
