package entities

import (
	"encoding/json"
	"testing"
)

func TestVideoJob_JSONMarshal(t *testing.T) {
	job := VideoJob{
		VideoName: "test-video.mp4",
		VideoPath: "/path/to/output",
		JobId:     "job-123",
	}

	jsonData, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("Failed to marshal VideoJob: %v", err)
	}

	expected := `{"videoUrl":"test-video.mp4","outputPath":"/path/to/output","jobId":"job-123"}`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON %s, got %s", expected, string(jsonData))
	}
}

func TestVideoJob_JSONUnmarshal(t *testing.T) {
	jsonData := `{"videoUrl":"test-video.mp4","outputPath":"/path/to/output","jobId":"job-123"}`

	var job VideoJob
	err := json.Unmarshal([]byte(jsonData), &job)
	if err != nil {
		t.Fatalf("Failed to unmarshal VideoJob: %v", err)
	}

	if job.VideoName != "test-video.mp4" {
		t.Errorf("Expected VideoName 'test-video.mp4', got '%s'", job.VideoName)
	}
	if job.VideoPath != "/path/to/output" {
		t.Errorf("Expected VideoPath '/path/to/output', got '%s'", job.VideoPath)
	}
	if job.JobId != "job-123" {
		t.Errorf("Expected JobId 'job-123', got '%s'", job.JobId)
	}
}

func TestVideoJob_EmptyValues(t *testing.T) {
	job := VideoJob{}

	if job.VideoName != "" {
		t.Errorf("Expected empty VideoName, got '%s'", job.VideoName)
	}
	if job.VideoPath != "" {
		t.Errorf("Expected empty VideoPath, got '%s'", job.VideoPath)
	}
	if job.JobId != "" {
		t.Errorf("Expected empty JobId, got '%s'", job.JobId)
	}
}

func TestVideoJob_WithSpecialCharacters(t *testing.T) {
	job := VideoJob{
		VideoName: "video with spaces & special chars!.mp4",
		VideoPath: "/path/with spaces/output",
		JobId:     "job-with-special-chars-123!@#",
	}

	jsonData, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("Failed to marshal VideoJob with special characters: %v", err)
	}

	var unmarshaled VideoJob
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal VideoJob with special characters: %v", err)
	}

	if unmarshaled.VideoName != job.VideoName {
		t.Errorf("VideoName mismatch after marshal/unmarshal")
	}
	if unmarshaled.VideoPath != job.VideoPath {
		t.Errorf("VideoPath mismatch after marshal/unmarshal")
	}
	if unmarshaled.JobId != job.JobId {
		t.Errorf("JobId mismatch after marshal/unmarshal")
	}
}
