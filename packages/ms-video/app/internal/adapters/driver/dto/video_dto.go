package dto

import "mime/multipart"

type UploadVideoInput struct {
	File      *multipart.FileHeader
	UserID    string
	UserEmail string
}

type UploadVideoOutput struct {
	VideoID      string `json:"video_id"`
	OriginalName string `json:"original_name"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type ListVideosOutput struct {
	Videos []VideoOutput `json:"videos"`
}

type VideoOutput struct {
	ID              string `json:"id"`
	OriginalName    string `json:"original_name"`
	Status          string `json:"status"`
	ProgressPercent int    `json:"progress_percent"`
	FileSize        int64  `json:"file_size"`
	ErrorMessage    string `json:"error_message,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type DownloadVideoOutput struct {
	PresignedURL string `json:"presigned_url"`
	VideoID      string `json:"video_id"`
	FileName     string `json:"file_name"`
	ExpiresIn    int    `json:"expires_in"`
}

type VideoProcessMessage struct {
	VideoID   string `json:"video_id"`
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	RawS3Key  string `json:"raw_s3_key"`
}
