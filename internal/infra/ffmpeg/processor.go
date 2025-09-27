package ffmpeg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
	var videoPath string
	var shouldCleanup bool

	if strings.HasPrefix(job.VideoPath, "https://") && strings.Contains(job.VideoPath, ".s3.") {
		localVideoPath := fmt.Sprintf("temp_video_%s.mp4", job.JobId)

		parts := strings.Split(job.VideoPath, "/")
		if len(parts) >= 4 {
			s3Key := strings.Join(parts[3:], "/")

			err := p.storage.Download(s3Key, localVideoPath)
			if err != nil {
				return &entities.ProcessingResult{
					Status: "failed",
					JobId:  job.JobId,
				}, fmt.Errorf("failed to download video from S3: %v", err)
			}
			videoPath = localVideoPath
			shouldCleanup = true
		} else {
			return &entities.ProcessingResult{
				Status: "failed",
				JobId:  job.JobId,
			}, fmt.Errorf("invalid S3 URL format: %s", job.VideoPath)
		}
	} else {
		videoPath = job.VideoPath
		shouldCleanup = false
	}

	if shouldCleanup {
		defer os.Remove(videoPath)
	}

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
		"-i", videoPath,
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
		log.Printf("Warning: Error removing frames directory %s: %v\n", outputDir, err)
	}

	return &entities.ProcessingResult{
		Status:     "completed",
		JobId:      job.JobId,
		OutputPath: storageResult.URL,
	}, nil
}
