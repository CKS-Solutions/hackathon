package entities

import (
	"testing"
	"time"
)

func TestNewVideo(t *testing.T) {
	userID := "user-123"
	userEmail := "user@example.com"
	originalName := "test-video.mp4"
	rawS3Key := "raw/user-123/test-video.mp4"
	fileSize := int64(1024000)

	video := NewVideo(userID, userEmail, originalName, rawS3Key, fileSize)

	if video.ID == "" {
		t.Error("expected video ID to be generated")
	}

	if video.UserID != userID {
		t.Errorf("expected UserID '%s', got '%s'", userID, video.UserID)
	}

	if video.UserEmail != userEmail {
		t.Errorf("expected UserEmail '%s', got '%s'", userEmail, video.UserEmail)
	}

	if video.OriginalName != originalName {
		t.Errorf("expected OriginalName '%s', got '%s'", originalName, video.OriginalName)
	}

	if video.RawS3Key != rawS3Key {
		t.Errorf("expected RawS3Key '%s', got '%s'", rawS3Key, video.RawS3Key)
	}

	if video.FileSize != fileSize {
		t.Errorf("expected FileSize %d, got %d", fileSize, video.FileSize)
	}

	if video.Status != VideoStatusPending {
		t.Errorf("expected Status '%s', got '%s'", VideoStatusPending, video.Status)
	}

	if video.ProgressPercent != 0 {
		t.Errorf("expected ProgressPercent 0, got %d", video.ProgressPercent)
	}

	if video.ProcessedS3Key != "" {
		t.Errorf("expected ProcessedS3Key to be empty, got '%s'", video.ProcessedS3Key)
	}

	if video.ErrorMessage != "" {
		t.Errorf("expected ErrorMessage to be empty, got '%s'", video.ErrorMessage)
	}

	if video.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if video.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	if !video.CreatedAt.Equal(video.UpdatedAt) {
		t.Error("expected CreatedAt and UpdatedAt to be equal on creation")
	}
}

func TestVideo_UpdateProgress(t *testing.T) {
	video := NewVideo("user-123", "user@example.com", "test.mp4", "raw/test.mp4", 1024)
	initialUpdatedAt := video.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name    string
		percent int
		status  VideoStatus
	}{
		{
			name:    "should update to processing with 25%",
			percent: 25,
			status:  VideoStatusProcessing,
		},
		{
			name:    "should update to processing with 50%",
			percent: 50,
			status:  VideoStatusProcessing,
		},
		{
			name:    "should update to processing with 75%",
			percent: 75,
			status:  VideoStatusProcessing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video.UpdateProgress(tt.percent, tt.status)

			if video.ProgressPercent != tt.percent {
				t.Errorf("expected ProgressPercent %d, got %d", tt.percent, video.ProgressPercent)
			}

			if video.Status != tt.status {
				t.Errorf("expected Status '%s', got '%s'", tt.status, video.Status)
			}

			if !video.UpdatedAt.After(initialUpdatedAt) {
				t.Error("expected UpdatedAt to be updated")
			}
		})
	}
}

func TestVideo_MarkAsCompleted(t *testing.T) {
	video := NewVideo("user-123", "user@example.com", "test.mp4", "raw/test.mp4", 1024)
	initialUpdatedAt := video.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	processedS3Key := "processed/user-123/test.zip"
	video.MarkAsCompleted(processedS3Key)

	if video.ProcessedS3Key != processedS3Key {
		t.Errorf("expected ProcessedS3Key '%s', got '%s'", processedS3Key, video.ProcessedS3Key)
	}

	if video.Status != VideoStatusCompleted {
		t.Errorf("expected Status '%s', got '%s'", VideoStatusCompleted, video.Status)
	}

	if video.ProgressPercent != 100 {
		t.Errorf("expected ProgressPercent 100, got %d", video.ProgressPercent)
	}

	if !video.UpdatedAt.After(initialUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestVideo_MarkAsFailed(t *testing.T) {
	video := NewVideo("user-123", "user@example.com", "test.mp4", "raw/test.mp4", 1024)
	initialUpdatedAt := video.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	errorMessage := "failed to process video: codec not supported"
	video.MarkAsFailed(errorMessage)

	if video.Status != VideoStatusFailed {
		t.Errorf("expected Status '%s', got '%s'", VideoStatusFailed, video.Status)
	}

	if video.ErrorMessage != errorMessage {
		t.Errorf("expected ErrorMessage '%s', got '%s'", errorMessage, video.ErrorMessage)
	}

	if !video.UpdatedAt.After(initialUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestVideoStatus_Constants(t *testing.T) {
	if VideoStatusPending != "pending" {
		t.Errorf("expected VideoStatusPending to be 'pending', got '%s'", VideoStatusPending)
	}

	if VideoStatusProcessing != "processing" {
		t.Errorf("expected VideoStatusProcessing to be 'processing', got '%s'", VideoStatusProcessing)
	}

	if VideoStatusCompleted != "completed" {
		t.Errorf("expected VideoStatusCompleted to be 'completed', got '%s'", VideoStatusCompleted)
	}

	if VideoStatusFailed != "failed" {
		t.Errorf("expected VideoStatusFailed to be 'failed', got '%s'", VideoStatusFailed)
	}
}
