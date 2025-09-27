package usecases

import (
	"encoding/json"
	"log"
	"upframer-worker/internal/domain/entities"
	"upframer-worker/internal/domain/services"
)

type ProcessVideoUseCase struct {
	processor services.VideoProcessor
	publisher services.Publisher
}

func NewProcessVideoUseCase(processor services.VideoProcessor, publisher services.Publisher) *ProcessVideoUseCase {
	return &ProcessVideoUseCase{
		processor: processor,
		publisher: publisher,
	}
}

func (p *ProcessVideoUseCase) Execute(messageRawData []byte) error {

	var job entities.VideoJob

	if err := json.Unmarshal(messageRawData, &job); err != nil {
		log.Printf("error parsing message: %v", err)
		return err
	}

	result, err := p.processor.ProcessVideo(&job)

	if err != nil {
		return err
	}

	if err := p.publisher.Publish("video-processing-result", result); err != nil {
		log.Printf("Error publishing result: %v", err)
		return err
	}

	return nil
}
