package entities

import (
	"encoding/json"
	"testing"
)

func TestProcessingResult_Creation(t *testing.T) {
	result := ProcessingResult{
		Status:     "success",
		OutputPath: "/path/to/processed/video.mp4",
		JobId:      "job-123",
	}

	if result.Status != "success" {
		t.Errorf("Expected Status 'success', got '%s'", result.Status)
	}
	if result.OutputPath != "/path/to/processed/video.mp4" {
		t.Errorf("Expected OutputPath '/path/to/processed/video.mp4', got '%s'", result.OutputPath)
	}
	if result.JobId != "job-123" {
		t.Errorf("Expected JobId 'job-123', got '%s'", result.JobId)
	}
}

func TestProcessingResult_JSONMarshal(t *testing.T) {
	result := ProcessingResult{
		Status:     "failed",
		OutputPath: "",
		JobId:      "job-456",
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ProcessingResult: %v", err)
	}

	expected := `{"Status":"failed","OutputPath":"","JobId":"job-456"}`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON %s, got %s", expected, string(jsonData))
	}
}

func TestProcessingResult_JSONUnmarshal(t *testing.T) {
	jsonData := `{"Status":"processing","OutputPath":"/tmp/video.mp4","JobId":"job-789"}`

	var result ProcessingResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProcessingResult: %v", err)
	}

	if result.Status != "processing" {
		t.Errorf("Expected Status 'processing', got '%s'", result.Status)
	}
	if result.OutputPath != "/tmp/video.mp4" {
		t.Errorf("Expected OutputPath '/tmp/video.mp4', got '%s'", result.OutputPath)
	}
	if result.JobId != "job-789" {
		t.Errorf("Expected JobId 'job-789', got '%s'", result.JobId)
	}
}

func TestProcessingResult_EmptyValues(t *testing.T) {
	result := ProcessingResult{}

	if result.Status != "" {
		t.Errorf("Expected empty Status, got '%s'", result.Status)
	}
	if result.OutputPath != "" {
		t.Errorf("Expected empty OutputPath, got '%s'", result.OutputPath)
	}
	if result.JobId != "" {
		t.Errorf("Expected empty JobId, got '%s'", result.JobId)
	}
}

func TestProcessingResult_StatusVariations(t *testing.T) {
	statuses := []string{"success", "failed", "processing", "queued", "error"}

	for _, status := range statuses {
		result := ProcessingResult{
			Status:     status,
			OutputPath: "/path/output.mp4",
			JobId:      "test-job",
		}

		if result.Status != status {
			t.Errorf("Expected Status '%s', got '%s'", status, result.Status)
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			t.Errorf("Failed to marshal ProcessingResult with status '%s': %v", status, err)
		}

		var unmarshaled ProcessingResult
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal ProcessingResult with status '%s': %v", status, err)
		}

		if unmarshaled.Status != status {
			t.Errorf("Status mismatch after marshal/unmarshal for status '%s'", status)
		}
	}
}
