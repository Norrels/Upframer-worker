package entities

type VideoJob struct {
	VideoName string `json:"videoUrl"`
	VideoPath string `json:"outputPath"`
	JobId     string `json:"jobId"`
}
