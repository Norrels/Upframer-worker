package errors

import "errors"

var (
	ErrFileNotFound        = errors.New("file not found in storage")
	ErrInvalidURLFormat    = errors.New("invalid URL format")
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrInvalidJobData      = errors.New("invalid job data")

	ErrStorageUnavailable  = errors.New("storage temporarily unavailable")
	ErrNetworkTimeout      = errors.New("network timeout")
	ErrFFmpegProcessing    = errors.New("ffmpeg processing error")
)

func IsPermanentError(err error) bool {
	return errors.Is(err, ErrFileNotFound) ||
		errors.Is(err, ErrInvalidURLFormat) ||
		errors.Is(err, ErrInvalidMessageFormat) ||
		errors.Is(err, ErrInvalidJobData)
}

func IsTemporaryError(err error) bool {
	return errors.Is(err, ErrStorageUnavailable) ||
		errors.Is(err, ErrNetworkTimeout) ||
		errors.Is(err, ErrFFmpegProcessing)
}
