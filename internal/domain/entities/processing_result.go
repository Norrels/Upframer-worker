package entities

type ProcessingResult struct {
	Success  bool
	Error    error
	ErrorMsg string
	OutputPath string
}
