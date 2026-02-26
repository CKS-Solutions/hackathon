package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/middleware"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/mocks"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

func TestVideoController_Upload_Success(t *testing.T) {
	// Create mocks
	videoRepo := &mocks.MockVideoRepository{
		SaveFunc: func(ctx context.Context, video *entities.Video) error {
			return nil
		},
	}

	storageService := &mocks.MockStorageService{
		UploadFunc: func(ctx context.Context, key string, data []byte, contentType string) error {
			return nil
		},
	}

	videoQueue := &mocks.MockVideoQueue{
		SendFunc: func(ctx context.Context, message dto.VideoProcessMessage) error {
			return nil
		},
	}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "test-video.mp4")
	part.Write([]byte("fake video content"))
	writer.Close()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/videos/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "user-123")
	ctx = context.WithValue(ctx, middleware.EmailContextKey, "user@example.com")
	req = req.WithContext(ctx)

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute
	err := controller.Upload(ctx, w, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response dto.UploadVideoOutput
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.VideoID == "" {
		t.Error("expected VideoID to be set")
	}

	if response.OriginalName != "test-video.mp4" {
		t.Errorf("expected OriginalName 'test-video.mp4', got '%s'", response.OriginalName)
	}
}

func TestVideoController_Upload_MethodNotAllowed(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos/upload", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "user-123")
	ctx = context.WithValue(ctx, middleware.EmailContextKey, "user@example.com")

	w := httptest.NewRecorder()

	err := controller.Upload(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for method not allowed, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, httpErr.StatusCode)
	}
}

func TestVideoController_Upload_MissingUserID(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "test-video.mp4")
	part.Write([]byte("fake video content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/videos/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Context without user ID
	ctx := req.Context()

	w := httptest.NewRecorder()

	err := controller.Upload(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for missing user ID, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

func TestVideoController_List_Success(t *testing.T) {
	userID := "user-123"
	videos := []*entities.Video{
		{
			ID:              "video-1",
			UserID:          userID,
			OriginalName:    "video1.mp4",
			Status:          entities.VideoStatusCompleted,
			ProgressPercent: 100,
			FileSize:        1024000,
		},
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			return videos, nil
		},
	}

	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)

	w := httptest.NewRecorder()

	err := controller.List(ctx, w, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response dto.ListVideosOutput
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Videos) != 1 {
		t.Errorf("expected 1 video, got %d", len(response.Videos))
	}
}

func TestVideoController_List_MethodNotAllowed(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodPost, "/videos", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "user-123")

	w := httptest.NewRecorder()

	err := controller.List(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for method not allowed, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, httpErr.StatusCode)
	}
}

func TestVideoController_List_MissingUserID(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	ctx := req.Context()

	w := httptest.NewRecorder()

	err := controller.List(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for missing user ID, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

func TestVideoController_Download_Success(t *testing.T) {
	userID := "user-123"
	videoID := "video-123"
	presignedURL := "https://s3.amazonaws.com/bucket/presigned-url"

	video := &entities.Video{
		ID:             videoID,
		UserID:         userID,
		OriginalName:   "test-video.mp4",
		ProcessedS3Key: "processed/user-123/test-video.zip",
		Status:         entities.VideoStatusCompleted,
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return video, nil
		},
	}

	storageService := &mocks.MockStorageService{
		GetPresignedURLFunc: func(ctx context.Context, key string, expirationMinutes int) (string, error) {
			return presignedURL, nil
		},
	}

	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos/download?id="+videoID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)

	w := httptest.NewRecorder()

	err := controller.Download(ctx, w, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response dto.DownloadVideoOutput
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.PresignedURL != presignedURL {
		t.Errorf("expected PresignedURL '%s', got '%s'", presignedURL, response.PresignedURL)
	}

	if response.VideoID != videoID {
		t.Errorf("expected VideoID '%s', got '%s'", videoID, response.VideoID)
	}
}

func TestVideoController_Download_MethodNotAllowed(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodPost, "/videos/download", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "user-123")

	w := httptest.NewRecorder()

	err := controller.Download(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for method not allowed, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, httpErr.StatusCode)
	}
}

func TestVideoController_Download_MissingVideoID(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos/download", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "user-123")

	w := httptest.NewRecorder()

	err := controller.Download(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for missing video ID, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, httpErr.StatusCode)
	}
}

func TestVideoController_Download_MissingUserID(t *testing.T) {
	videoRepo := &mocks.MockVideoRepository{}
	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos/download?id=video-123", nil)
	ctx := req.Context()

	w := httptest.NewRecorder()

	err := controller.Download(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for missing user ID, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

func TestVideoController_Download_VideoNotFound(t *testing.T) {
	userID := "user-123"
	videoID := "non-existent-video"

	videoRepo := &mocks.MockVideoRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entities.Video, error) {
			return nil, errors.New("video not found")
		},
	}

	storageService := &mocks.MockStorageService{}
	videoQueue := &mocks.MockVideoQueue{}

	uploadUsecase := usecases.NewUploadVideoUsecase(videoRepo, storageService, videoQueue)
	listUsecase := usecases.NewListVideosUsecase(videoRepo)
	downloadUsecase := usecases.NewDownloadVideoUsecase(videoRepo, storageService)

	controller := NewVideoController(uploadUsecase, listUsecase, downloadUsecase)

	req := httptest.NewRequest(http.MethodGet, "/videos/download?id="+videoID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)

	w := httptest.NewRecorder()

	err := controller.Download(ctx, w, req)

	if err == nil {
		t.Fatal("expected error for video not found, got nil")
	}

	httpErr, ok := err.(*utils.HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, httpErr.StatusCode)
	}
}
