package entities

import (
	"time"

	"github.com/google/uuid"
)

type VideoStatus string

const (
	VideoStatusPending    VideoStatus = "pending"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusCompleted  VideoStatus = "completed"
	VideoStatusFailed     VideoStatus = "failed"
)

type Video struct {
	ID              string      `json:"id" dynamodbav:"id"`
	UserID          string      `json:"user_id" dynamodbav:"user_id"`
	UserEmail       string      `json:"user_email" dynamodbav:"user_email"`
	OriginalName    string      `json:"original_name" dynamodbav:"original_name"`
	RawS3Key        string      `json:"raw_s3_key" dynamodbav:"raw_s3_key"`
	ProcessedS3Key  string      `json:"processed_s3_key,omitempty" dynamodbav:"processed_s3_key"`
	Status          VideoStatus `json:"status" dynamodbav:"status"`
	ProgressPercent int         `json:"progress_percent" dynamodbav:"progress_percent"`
	ErrorMessage    string      `json:"error_message,omitempty" dynamodbav:"error_message"`
	FileSize        int64       `json:"file_size" dynamodbav:"file_size"`
	CreatedAt       time.Time   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" dynamodbav:"updated_at"`
}

func NewVideo(userID, userEmail, originalName, rawS3Key string, fileSize int64) *Video {
	now := time.Now()
	return &Video{
		ID:              uuid.NewString(),
		UserID:          userID,
		UserEmail:       userEmail,
		OriginalName:    originalName,
		RawS3Key:        rawS3Key,
		Status:          VideoStatusPending,
		ProgressPercent: 0,
		FileSize:        fileSize,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (v *Video) UpdateProgress(percent int, status VideoStatus) {
	v.ProgressPercent = percent
	v.Status = status
	v.UpdatedAt = time.Now()
}

func (v *Video) MarkAsCompleted(processedS3Key string) {
	v.ProcessedS3Key = processedS3Key
	v.Status = VideoStatusCompleted
	v.ProgressPercent = 100
	v.UpdatedAt = time.Now()
}

func (v *Video) MarkAsFailed(errorMessage string) {
	v.Status = VideoStatusFailed
	v.ErrorMessage = errorMessage
	v.UpdatedAt = time.Now()
}
