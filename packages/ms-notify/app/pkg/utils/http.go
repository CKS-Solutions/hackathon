package utils

import (
	"errors"
	"net/http"
)

type HttpError struct {
	error
	StatusCode int
}

func NewHttpError(statusCode int, message string) *HttpError {
	return &HttpError{
		error:      errors.New(message),
		StatusCode: statusCode,
	}
}

func HTTPBadRequest(message string) *HttpError {
	return NewHttpError(http.StatusBadRequest, message)
}
