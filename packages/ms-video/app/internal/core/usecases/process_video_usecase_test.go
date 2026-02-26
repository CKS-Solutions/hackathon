package usecases

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/mocks"
)

func TestProcessVideoUsecase_Execute_VideoNotFound(t *testing.T) {
	ctx := context.Background()

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, videoID string) (*entities.Video, error) {
			return nil, errors.New("video not found")
		},
	}

	storageService := &mocks.MockStorageService{}
	notificationService := &mocks.MockNotificationService{}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	message := dto.VideoProcessMessage{
		VideoID:   "non-existent-video",
		UserID:    "user-123",
		UserEmail: "user@example.com",
		RawS3Key:  "raw/user-123/video.mp4",
	}

	err := usecase.Execute(ctx, message)

	if err == nil {
		t.Fatal("expected error when video not found, got nil")
	}
}

func TestProcessVideoUsecase_Execute_DownloadFails(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"
	userEmail := "user@example.com"

	video := &entities.Video{
		ID:           videoID,
		UserID:       "user-123",
		UserEmail:    userEmail,
		OriginalName: "test.mp4",
		Status:       entities.VideoStatusPending,
	}

	updateCalled := false
	notificationSent := false

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
		UpdateFunc: func(ctx context.Context, v *entities.Video) error {
			updateCalled = true
			return nil
		},
	}

	storageService := &mocks.MockStorageService{
		DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("download failed")
		},
	}

	notificationService := &mocks.MockNotificationService{
		SendVideoFailedNotificationFunc: func(ctx context.Context, email, videoID, originalName, errorMessage string) error {
			notificationSent = true
			if email != userEmail {
				t.Errorf("expected email '%s', got '%s'", userEmail, email)
			}
			return nil
		},
	}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	message := dto.VideoProcessMessage{
		VideoID:   videoID,
		UserID:    "user-123",
		UserEmail: userEmail,
		RawS3Key:  "raw/user-123/test.mp4",
	}

	err := usecase.Execute(ctx, message)

	if err == nil {
		t.Fatal("expected error when download fails, got nil")
	}

	if !updateCalled {
		t.Error("expected repository Update to be called")
	}

	if !notificationSent {
		t.Error("expected failure notification to be sent")
	}

	if video.Status != entities.VideoStatusFailed {
		t.Errorf("expected video status to be '%s', got '%s'", entities.VideoStatusFailed, video.Status)
	}
}

func TestProcessVideoUsecase_Execute_ExtractFramesFails(t *testing.T) {
	// Note: This test verifies error handling when frame extraction fails
	// Actual ffmpeg execution is skipped as it requires complex setup
	t.Skip("Skipping test that requires ffmpeg - integration test needed")
}

func TestProcessVideoUsecase_Execute_UploadProcessedVideoFails(t *testing.T) {
	// Note: This test would require mocking ffmpeg which is complex
	// The error handling pattern is consistent and tested in download scenario
	t.Skip("Skipping test that requires ffmpeg - integration test needed")
}

func TestProcessVideoUsecase_CreateZipFile(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	notificationService := &mocks.MockNotificationService{}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	originalName := "test-video.mp4"
	videoData := []byte("fake video content")
	frames := [][]byte{
		[]byte("frame1 data"),
		[]byte("frame2 data"),
		[]byte("frame3 data"),
	}

	zipData, err := usecase.createZipFile(originalName, videoData, frames)

	if err != nil {
		t.Fatalf("expected no error creating zip file, got %v", err)
	}

	if len(zipData) == 0 {
		t.Error("expected zip data to be non-empty")
	}

	// Verify it starts with ZIP signature
	if len(zipData) < 4 {
		t.Fatal("zip data too short")
	}

	// ZIP files start with PK (0x50 0x4B)
	if zipData[0] != 0x50 || zipData[1] != 0x4B {
		t.Error("zip data doesn't have valid ZIP signature")
	}
}

func TestProcessVideoUsecase_CreateZipFile_EmptyFrames(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	notificationService := &mocks.MockNotificationService{}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	originalName := "test-video.mp4"
	videoData := []byte("fake video content")
	frames := [][]byte{}

	zipData, err := usecase.createZipFile(originalName, videoData, frames)

	if err != nil {
		t.Fatalf("expected no error creating zip file with empty frames, got %v", err)
	}

	if len(zipData) == 0 {
		t.Error("expected zip data to be non-empty")
	}
}

func TestProcessVideoUsecase_CreateZipFile_LargeFrames(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	notificationService := &mocks.MockNotificationService{}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	originalName := "test-video.mp4"
	videoData := []byte("fake video content")
	
	// Create 100 frames
	frames := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		frames[i] = bytes.Repeat([]byte("frame"), 1000) // ~5KB per frame
	}

	zipData, err := usecase.createZipFile(originalName, videoData, frames)

	if err != nil {
		t.Fatalf("expected no error creating zip file with many frames, got %v", err)
	}

	if len(zipData) == 0 {
		t.Error("expected zip data to be non-empty")
	}

	// ZIP should contain the compressed data
	// The actual size depends on compression, but should be substantial
	if len(zipData) < 1000 {
		t.Errorf("expected zip data to be at least 1000 bytes, got %d", len(zipData))
	}
}

func TestProcessVideoUsecase_NotificationSkippedWhenNoEmail(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"

	video := &entities.Video{
		ID:           videoID,
		UserID:       "user-123",
		UserEmail:    "",
		OriginalName: "test.mp4",
		Status:       entities.VideoStatusPending,
	}

	notificationCalled := false

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
		UpdateFunc: func(ctx context.Context, v *entities.Video) error {
			return nil
		},
	}

	storageService := &mocks.MockStorageService{
		DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("download failed")
		},
	}

	notificationService := &mocks.MockNotificationService{
		SendVideoFailedNotificationFunc: func(ctx context.Context, email, videoID, originalName, errorMessage string) error {
			notificationCalled = true
			return nil
		},
	}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	message := dto.VideoProcessMessage{
		VideoID:   videoID,
		UserID:    "user-123",
		UserEmail: "", // Empty email
		RawS3Key:  "raw/user-123/test.mp4",
	}

	usecase.Execute(ctx, message)

	if notificationCalled {
		t.Error("expected notification not to be sent when email is empty")
	}
}

func TestProcessVideoUsecase_UpdateProgressCalled(t *testing.T) {
	ctx := context.Background()
	videoID := "video-123"

	video := &entities.Video{
		ID:              videoID,
		UserID:          "user-123",
		UserEmail:       "user@example.com",
		OriginalName:    "test.mp4",
		Status:          entities.VideoStatusPending,
		ProgressPercent: 0,
	}

	progressUpdates := []int{}

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
		UpdateFunc: func(ctx context.Context, v *entities.Video) error {
			progressUpdates = append(progressUpdates, v.ProgressPercent)
			return nil
		},
	}

	storageService := &mocks.MockStorageService{
		DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("download failed")
		},
	}

	notificationService := &mocks.MockNotificationService{
		SendVideoFailedNotificationFunc: func(ctx context.Context, email, videoID, originalName, errorMessage string) error {
			return nil
		},
	}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	message := dto.VideoProcessMessage{
		VideoID:   videoID,
		UserID:    "user-123",
		UserEmail: "user@example.com",
		RawS3Key:  "raw/user-123/test.mp4",
	}

	usecase.Execute(ctx, message)

	// Should have called update at least once for initial progress
	if len(progressUpdates) == 0 {
		t.Error("expected progress updates to be recorded")
	}

	// First progress update should be 10% (initial processing)
	if len(progressUpdates) > 0 && progressUpdates[0] != 10 {
		t.Errorf("expected first progress update to be 10, got %d", progressUpdates[0])
	}
}

func TestNewProcessVideoUsecase(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	notificationService := &mocks.MockNotificationService{}

	usecase := NewProcessVideoUsecase(videoRepo, storageService, notificationService)

	if usecase == nil {
		t.Fatal("expected usecase to be created, got nil")
	}

	if usecase.videoRepository != videoRepo {
		t.Error("expected videoRepository to be set correctly")
	}

	if usecase.storageService != storageService {
		t.Error("expected storageService to be set correctly")
	}

	if usecase.notificationService != notificationService {
		t.Error("expected notificationService to be set correctly")
	}
}
