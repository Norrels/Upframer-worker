package usecases

import (
	"encoding/json"
	"fmt"
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
		fmt.Printf("error parsing message: %v\n", err)
		return err
	}

	result, err := p.processor.ProcessVideo(&job)

	if err != nil {
		return err
	}

	if err := p.publisher.Publish("video-processing-result", result); err != nil {
		fmt.Printf("Error publishing result: %v\n", err)
		return err
	}

	return nil
}
