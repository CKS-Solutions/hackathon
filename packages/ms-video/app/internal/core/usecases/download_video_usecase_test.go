package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/mocks"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

func TestDownloadVideoUsecase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"
	userID := "user-123"
	processedS3Key := "processed/user-123/test-video.zip"
	presignedURL := "https://s3.amazonaws.com/bucket/presigned-url"

	video := &entities.Video{
		ID:             videoID,
		UserID:         userID,
		OriginalName:   "test-video.mp4",
		ProcessedS3Key: processedS3Key,
		Status:         entities.VideoStatusCompleted,
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			if id != videoID {
				t.Errorf("expected videoID '%s', got '%s'", videoID, id)
			}
			return video, nil
		},
	}

	storageService := &mocks.MockStorageService{
		GetPresignedURLFunc: func(ctx context.Context, key string, expirationMinutes int) (string, error) {
			if key != processedS3Key {
				t.Errorf("expected key '%s', got '%s'", processedS3Key, key)
			}
			if expirationMinutes != 15 {
				t.Errorf("expected expiration 15 minutes, got %d", expirationMinutes)
			}
			return presignedURL, nil
		},
	}

	usecase := NewDownloadVideoUsecase(videoRepo, storageService)

	output, err := usecase.Execute(ctx, videoID, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output == nil {
		t.Fatal("expected output, got nil")
	}

	if output.PresignedURL != presignedURL {
		t.Errorf("expected PresignedURL '%s', got '%s'", presignedURL, output.PresignedURL)
	}

	if output.VideoID != videoID {
		t.Errorf("expected VideoID '%s', got '%s'", videoID, output.VideoID)
	}

	if output.FileName != "test-video.mp4.zip" {
		t.Errorf("expected FileName 'test-video.mp4.zip', got '%s'", output.FileName)
	}

	expectedExpiresIn := 15 * 60 // 15 minutes in seconds
	if output.ExpiresIn != expectedExpiresIn {
		t.Errorf("expected ExpiresIn %d, got %d", expectedExpiresIn, output.ExpiresIn)
	}
}

func TestDownloadVideoUsecase_Execute_VideoNotFound(t *testing.T) {
	ctx := context.Background()
	videoID := "non-existent-video"
	userID := "user-123"

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return nil, errors.New("video not found")
		},
	}

	storageService := &mocks.MockStorageService{}

	usecase := NewDownloadVideoUsecase(videoRepo, storageService)

	output, err := usecase.Execute(ctx, videoID, userID)

	if err == nil {
		t.Fatal("expected error for video not found, got nil")
	}

	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", httpErr.StatusCode)
	}
}

func TestDownloadVideoUsecase_Execute_UnauthorizedAccess(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"
	ownerUserID := "user-123"
	differentUserID := "user-456"

	video := &entities.Video{
		ID:             videoID,
		UserID:         ownerUserID,
		OriginalName:   "test-video.mp4",
		ProcessedS3Key: "processed/user-123/test-video.zip",
		Status:         entities.VideoStatusCompleted,
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
	}

	storageService := &mocks.MockStorageService{}

	usecase := NewDownloadVideoUsecase(videoRepo, storageService)

	output, err := usecase.Execute(ctx, videoID, differentUserID)

	if err == nil {
		t.Fatal("expected error for unauthorized access, got nil")
	}

	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", httpErr.StatusCode)
	}
}

func TestDownloadVideoUsecase_Execute_VideoNotReady(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"
	userID := "user-123"

	tests := []struct {
		name   string
		status entities.VideoStatus
	}{
		{
			name:   "should reject pending video",
			status: entities.VideoStatusPending,
		},
		{
			name:   "should reject processing video",
			status: entities.VideoStatusProcessing,
		},
		{
			name:   "should reject failed video",
			status: entities.VideoStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video := &entities.Video{
				ID:             videoID,
				UserID:         userID,
				OriginalName:   "test-video.mp4",
				ProcessedS3Key: "processed/user-123/test-video.zip",
				Status:         tt.status,
			}

			videoRepo := &mocks.MockVideoRepository{
				FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
					return video, nil
				},
			}

			storageService := &mocks.MockStorageService{}

			usecase := NewDownloadVideoUsecase(videoRepo, storageService)

			output, err := usecase.Execute(ctx, videoID, userID)

			if err == nil {
				t.Fatal("expected error for video not ready, got nil")
			}

			if output != nil {
				t.Errorf("expected nil output, got %v", output)
			}

			httpErr, ok := err.(*utils.HttpError)
			if !ok {
				t.Fatalf("expected HttpError, got %T", err)
			}

			if httpErr.StatusCode != 400 {
				t.Errorf("expected status code 400, got %d", httpErr.StatusCode)
			}
		})
	}
}

func TestDownloadVideoUsecase_Execute_PresignedURLGenerationFails(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"
	userID := "user-123"
	processedS3Key := "processed/user-123/test-video.zip"

	video := &entities.Video{
		ID:             videoID,
		UserID:         userID,
		OriginalName:   "test-video.mp4",
		ProcessedS3Key: processedS3Key,
		Status:         entities.VideoStatusCompleted,
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
	}

	storageService := &mocks.MockStorageService{
		GetPresignedURLFunc: func(ctx context.Context, key string, expirationMinutes int) (string, error) {
			return "", errors.New("failed to generate presigned URL")
		},
	}

	usecase := NewDownloadVideoUsecase(videoRepo, storageService)

	output, err := usecase.Execute(ctx, videoID, userID)

	if err == nil {
		t.Fatal("expected error when presigned URL generation fails, got nil")
	}

	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != 500 {
		t.Errorf("expected status code 500, got %d", httpErr.StatusCode)
	}
}
