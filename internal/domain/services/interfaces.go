package services

import "upframer-worker/internal/domain/entities"

type VideoProcessor interface {
	ProcessVideo(job *entities.VideoJob) (*entities.ProcessingResult, error)
}

type Publisher interface {
	Publish(queueName string, result *entities.ProcessingResult) error
}
