package utils

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	error
	StatusCode int
}

func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		error:      errors.New(message),
		StatusCode: statusCode,
	}
}

func HTTPBadRequest(message string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message)
}

func HTTPUnauthorized(message string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message)
}

func HTTPNotFound(message string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message)
}

func HTTPConflict(message string) *HTTPError {
	return NewHTTPError(http.StatusConflict, message)
}

func HTTPMethodNotAllowed(message string) *HTTPError {
	return NewHTTPError(http.StatusMethodNotAllowed, message)
}

func HTTPInternalServerError(message string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message)
}
