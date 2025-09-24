package ffmpeg

import (
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
			Status: "failed",
			JobId:  job.JobId,
		}, err
	}

	// TODO
	// Logar isso Error:  fmt.Errorf("error creating folder: %v", err),

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
			Status:     "completed",
			JobId:      job.JobId,
			OutputPath: outputDir,
		}, nil
	}

	return &entities.ProcessingResult{
		Status: "failed",
		JobId:  job.JobId,
	}, nil
}
