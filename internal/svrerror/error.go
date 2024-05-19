package svrerror

// Error is a custom error type used to wrap errors with a status code.
type Error struct {
	Message    string
	StatusCode int
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string, statusCode int) error {
	return &Error{
		Message:    message,
		StatusCode: statusCode,
	}
}
