package utils

import "fmt"

type HttpError struct {
	StatusCode int
	Message    string
}

func (e *HttpError) Error() string {
	return e.Message
}

func NewHttpError(statusCode int, message string) *HttpError {
	return &HttpError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func NewBadRequestError(message string) *HttpError {
	return NewHttpError(400, message)
}

func NewUnauthorizedError(message string) *HttpError {
	return NewHttpError(401, message)
}

func NewNotFoundError(message string) *HttpError {
	return NewHttpError(404, message)
}

func NewInternalServerError(message string) *HttpError {
	return NewHttpError(500, message)
}

func NewValidationError(field string) *HttpError {
	return NewBadRequestError(fmt.Sprintf("invalid or missing field: %s", field))
}
