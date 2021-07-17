package clierror

// SanitizedError allows errors to be wrapped with a sanitized message for sending upstream
type SanitizedError struct {
	sanitizedMsg string
	err          error
}

func (e *SanitizedError) Error() string {
	return e.err.Error()
}

func (e *SanitizedError) SanitizedError() string {
	return e.sanitizedMsg
}

func NewSanitizedError(err error, sanitizedMsg string) *SanitizedError {
	return &SanitizedError{
		sanitizedMsg: sanitizedMsg,
		err:          err,
	}
}
