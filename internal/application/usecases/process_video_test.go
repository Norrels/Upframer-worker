package usecases

import (
	"encoding/json"
	"errors"
	"testing"
	"upframer-worker/internal/domain/entities"
)

type MockVideoProcessor struct {
	processVideoFunc func(job *entities.VideoJob) (*entities.ProcessingResult, error)
}

func (m *MockVideoProcessor) ProcessVideo(job *entities.VideoJob) (*entities.ProcessingResult, error) {
	if m.processVideoFunc != nil {
		return m.processVideoFunc(job)
	}
	return &entities.ProcessingResult{
		Status:     "success",
		OutputPath: "/path/to/output.mp4",
		JobId:      job.JobId,
	}, nil
}

type MockPublisher struct {
	publishFunc func(queueName string, result *entities.ProcessingResult) error
}

func (m *MockPublisher) Publish(queueName string, result *entities.ProcessingResult) error {
	if m.publishFunc != nil {
		return m.publishFunc(queueName, result)
	}
	return nil
}

func TestNewProcessVideoUseCase(t *testing.T) {
	processor := &MockVideoProcessor{}
	publisher := &MockPublisher{}

	useCase := NewProcessVideoUseCase(processor, publisher)

	if useCase == nil {
		t.Fatal("Expected non-nil ProcessVideoUseCase")
	}
	if useCase.processor != processor {
		t.Error("Expected processor to be set correctly")
	}
	if useCase.publisher != publisher {
		t.Error("Expected publisher to be set correctly")
	}
}

func TestProcessVideoUseCase_Execute_Success(t *testing.T) {
	processor := &MockVideoProcessor{}
	publisher := &MockPublisher{}
	useCase := NewProcessVideoUseCase(processor, publisher)

	job := entities.VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	messageData, _ := json.Marshal(job)

	err := useCase.Execute(messageData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestProcessVideoUseCase_Execute_InvalidJSON(t *testing.T) {
	processor := &MockVideoProcessor{}
	publisher := &MockPublisher{}
	useCase := NewProcessVideoUseCase(processor, publisher)

	invalidJSON := []byte(`{"invalid json}`)

	err := useCase.Execute(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestProcessVideoUseCase_Execute_ProcessorError(t *testing.T) {
	processor := &MockVideoProcessor{
		processVideoFunc: func(job *entities.VideoJob) (*entities.ProcessingResult, error) {
			return nil, errors.New("processor error")
		},
	}
	publisher := &MockPublisher{}
	useCase := NewProcessVideoUseCase(processor, publisher)

	job := entities.VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	messageData, _ := json.Marshal(job)

	err := useCase.Execute(messageData)
	if err == nil {
		t.Error("Expected error from processor")
	}
	if err.Error() != "processor error" {
		t.Errorf("Expected 'processor error', got '%s'", err.Error())
	}
}

func TestProcessVideoUseCase_Execute_PublisherError(t *testing.T) {
	processor := &MockVideoProcessor{}
	publisher := &MockPublisher{
		publishFunc: func(queueName string, result *entities.ProcessingResult) error {
			return errors.New("publisher error")
		},
	}
	useCase := NewProcessVideoUseCase(processor, publisher)

	job := entities.VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	messageData, _ := json.Marshal(job)

	err := useCase.Execute(messageData)
	if err == nil {
		t.Error("Expected error from publisher")
	}
	if err.Error() != "publisher error" {
		t.Errorf("Expected 'publisher error', got '%s'", err.Error())
	}
}

func TestProcessVideoUseCase_Execute_CorrectQueueName(t *testing.T) {
	processor := &MockVideoProcessor{}

	var capturedQueueName string
	publisher := &MockPublisher{
		publishFunc: func(queueName string, result *entities.ProcessingResult) error {
			capturedQueueName = queueName
			return nil
		},
	}

	useCase := NewProcessVideoUseCase(processor, publisher)

	job := entities.VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	messageData, _ := json.Marshal(job)

	err := useCase.Execute(messageData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if capturedQueueName != "video-processing-result" {
		t.Errorf("Expected queue name 'video-processing-result', got '%s'", capturedQueueName)
	}
}

func TestProcessVideoUseCase_Execute_ProcessorReceivesCorrectJob(t *testing.T) {
	var capturedJob *entities.VideoJob
	processor := &MockVideoProcessor{
		processVideoFunc: func(job *entities.VideoJob) (*entities.ProcessingResult, error) {
			capturedJob = job
			return &entities.ProcessingResult{
				Status:     "success",
				OutputPath: "/output.mp4",
				JobId:      job.JobId,
			}, nil
		},
	}
	publisher := &MockPublisher{}
	useCase := NewProcessVideoUseCase(processor, publisher)

	expectedJob := entities.VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	messageData, _ := json.Marshal(expectedJob)

	err := useCase.Execute(messageData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if capturedJob == nil {
		t.Fatal("Expected job to be captured")
	}
	if capturedJob.VideoName != expectedJob.VideoName {
		t.Errorf("Expected VideoName '%s', got '%s'", expectedJob.VideoName, capturedJob.VideoName)
	}
	if capturedJob.VideoPath != expectedJob.VideoPath {
		t.Errorf("Expected VideoPath '%s', got '%s'", expectedJob.VideoPath, capturedJob.VideoPath)
	}
	if capturedJob.JobId != expectedJob.JobId {
		t.Errorf("Expected JobId '%s', got '%s'", expectedJob.JobId, capturedJob.JobId)
	}
}