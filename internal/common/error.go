package common

type RequestError struct {
	StatusCode int
	Message    string
}

func (e *RequestError) Error() string {
	return e.Message
}

// NewRequestError construit une erreur personnalisée
func NewRequestError(statusCode int, msg string) *RequestError {
	return &RequestError{
		StatusCode: statusCode,
		Message:    msg,
	}
}

func NewForbiddenError(msg string) *RequestError {
	return NewRequestError(403, msg)
}

func NewNotFoundError(msg string) *RequestError {
	return NewRequestError(404, msg)
}

func NewUnauthorizedError(msg string) *RequestError {
	return NewRequestError(401, msg)
}
