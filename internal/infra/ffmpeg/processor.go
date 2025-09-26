package ffmpeg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"upframer-worker/internal/domain/entities"
	"upframer-worker/internal/domain/ports"
)

type FFmpegProcessor struct {
	storage ports.Storage
}

func NewFFmpegProcessor(storage ports.Storage) *FFmpegProcessor {
	return &FFmpegProcessor{
		storage: storage,
	}
}

func (p *FFmpegProcessor) ProcessVideo(job *entities.VideoJob) (*entities.ProcessingResult, error) {
	outputDir := "frames"

	err := os.MkdirAll(outputDir, 0755)

	if err != nil {
		log.Fatal("Error creating directory: ", err)
		return &entities.ProcessingResult{
			Status: "failed",
			JobId:  job.JobId,
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
			Status: "failed",
			JobId:  job.JobId,
		}, err
	}

	zipFileName := fmt.Sprintf("frames_%s.zip", job.JobId)
	fmt.Printf("Creating ZIP file: %s\n", zipFileName)

	storageResult, err := p.storage.StoreZip(outputDir, zipFileName)
	if err != nil {
		log.Fatal("Error storing ZIP: ", err)
		return &entities.ProcessingResult{
			Status: "failed",
			JobId:  job.JobId,
		}, err
	}

	err = os.RemoveAll(outputDir)
	if err != nil {
		fmt.Printf("Warning: Error removing frames directory %s: %v\n", outputDir, err)
	}

	return &entities.ProcessingResult{
		Status:     "completed",
		JobId:      job.JobId,
		OutputPath: storageResult.Path,
	}, nil
}
