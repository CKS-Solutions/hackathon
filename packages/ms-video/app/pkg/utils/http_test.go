package utils

import (
	"testing"
)

func TestNewHttpError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "should create http error with custom status code",
			statusCode: 404,
			message:    "resource not found",
		},
		{
			name:       "should create http error with 500 status code",
			statusCode: 500,
			message:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHttpError(tt.statusCode, tt.message)

			if err.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, err.StatusCode)
			}

			if err.Message != tt.message {
				t.Errorf("expected message '%s', got '%s'", tt.message, err.Message)
			}

			if err.Error() != tt.message {
				t.Errorf("expected Error() to return '%s', got '%s'", tt.message, err.Error())
			}
		})
	}
}

func TestNewBadRequestError(t *testing.T) {
	message := "invalid request"
	err := NewBadRequestError(message)

	if err.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", err.StatusCode)
	}

	if err.Message != message {
		t.Errorf("expected message '%s', got '%s'", message, err.Message)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	message := "unauthorized access"
	err := NewUnauthorizedError(message)

	if err.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", err.StatusCode)
	}

	if err.Message != message {
		t.Errorf("expected message '%s', got '%s'", message, err.Message)
	}
}

func TestNewNotFoundError(t *testing.T) {
	message := "resource not found"
	err := NewNotFoundError(message)

	if err.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", err.StatusCode)
	}

	if err.Message != message {
		t.Errorf("expected message '%s', got '%s'", message, err.Message)
	}
}

func TestNewInternalServerError(t *testing.T) {
	message := "something went wrong"
	err := NewInternalServerError(message)

	if err.StatusCode != 500 {
		t.Errorf("expected status code 500, got %d", err.StatusCode)
	}

	if err.Message != message {
		t.Errorf("expected message '%s', got '%s'", message, err.Message)
	}
}

func TestNewValidationError(t *testing.T) {
	field := "email"
	err := NewValidationError(field)

	if err.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", err.StatusCode)
	}

	expectedMessage := "invalid or missing field: email"
	if err.Message != expectedMessage {
		t.Errorf("expected message '%s', got '%s'", expectedMessage, err.Message)
	}
}
