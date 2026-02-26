package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/mocks"
)

func TestListVideosUsecase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	now := time.Now()
	videos := []*entities.Video{
		{
			ID:              "video-1",
			UserID:          userID,
			OriginalName:    "video1.mp4",
			Status:          entities.VideoStatusCompleted,
			ProgressPercent: 100,
			FileSize:        1024000,
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-1 * time.Hour),
		},
		{
			ID:              "video-2",
			UserID:          userID,
			OriginalName:    "video2.mp4",
			Status:          entities.VideoStatusProcessing,
			ProgressPercent: 50,
			FileSize:        2048000,
			CreatedAt:       now.Add(-1 * time.Hour),
			UpdatedAt:       now.Add(-30 * time.Minute),
		},
		{
			ID:              "video-3",
			UserID:          userID,
			OriginalName:    "video3.mp4",
			Status:          entities.VideoStatusPending,
			ProgressPercent: 0,
			FileSize:        512000,
			CreatedAt:       now.Add(-10 * time.Minute),
			UpdatedAt:       now.Add(-10 * time.Minute),
		},
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			if uid != userID {
				t.Errorf("expected userID '%s', got '%s'", userID, uid)
			}
			return videos, nil
		},
	}

	usecase := NewListVideosUsecase(videoRepo)

	output, err := usecase.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output == nil {
		t.Fatal("expected output, got nil")
	}

	if len(output.Videos) != len(videos) {
		t.Fatalf("expected %d videos, got %d", len(videos), len(output.Videos))
	}

	for i, videoOutput := range output.Videos {
		expectedVideo := videos[i]

		if videoOutput.ID != expectedVideo.ID {
			t.Errorf("video[%d]: expected ID '%s', got '%s'", i, expectedVideo.ID, videoOutput.ID)
		}

		if videoOutput.OriginalName != expectedVideo.OriginalName {
			t.Errorf("video[%d]: expected OriginalName '%s', got '%s'", i, expectedVideo.OriginalName, videoOutput.OriginalName)
		}

		if videoOutput.Status != string(expectedVideo.Status) {
			t.Errorf("video[%d]: expected Status '%s', got '%s'", i, expectedVideo.Status, videoOutput.Status)
		}

		if videoOutput.ProgressPercent != expectedVideo.ProgressPercent {
			t.Errorf("video[%d]: expected ProgressPercent %d, got %d", i, expectedVideo.ProgressPercent, videoOutput.ProgressPercent)
		}

		if videoOutput.FileSize != expectedVideo.FileSize {
			t.Errorf("video[%d]: expected FileSize %d, got %d", i, expectedVideo.FileSize, videoOutput.FileSize)
		}

		if videoOutput.ErrorMessage != expectedVideo.ErrorMessage {
			t.Errorf("video[%d]: expected ErrorMessage '%s', got '%s'", i, expectedVideo.ErrorMessage, videoOutput.ErrorMessage)
		}

		expectedCreatedAt := expectedVideo.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		if videoOutput.CreatedAt != expectedCreatedAt {
			t.Errorf("video[%d]: expected CreatedAt '%s', got '%s'", i, expectedCreatedAt, videoOutput.CreatedAt)
		}

		expectedUpdatedAt := expectedVideo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		if videoOutput.UpdatedAt != expectedUpdatedAt {
			t.Errorf("video[%d]: expected UpdatedAt '%s', got '%s'", i, expectedUpdatedAt, videoOutput.UpdatedAt)
		}
	}
}

func TestListVideosUsecase_Execute_EmptyList(t *testing.T) {
	ctx := context.Background()
	userID := "user-with-no-videos"

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			return []*entities.Video{}, nil
		},
	}

	usecase := NewListVideosUsecase(videoRepo)

	output, err := usecase.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output == nil {
		t.Fatal("expected output, got nil")
	}

	if len(output.Videos) != 0 {
		t.Errorf("expected empty videos list, got %d videos", len(output.Videos))
	}
}

func TestListVideosUsecase_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			return nil, errors.New("database connection failed")
		},
	}

	usecase := NewListVideosUsecase(videoRepo)

	output, err := usecase.Execute(ctx, userID)

	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}

	if output != nil {
		t.Errorf("expected nil output, got %v", output)
	}
}

func TestListVideosUsecase_Execute_WithFailedVideo(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	now := time.Now()
	errorMessage := "codec not supported"
	videos := []*entities.Video{
		{
			ID:              "video-1",
			UserID:          userID,
			OriginalName:    "failed-video.mp4",
			Status:          entities.VideoStatusFailed,
			ProgressPercent: 30,
			FileSize:        1024000,
			ErrorMessage:    errorMessage,
			CreatedAt:       now.Add(-1 * time.Hour),
			UpdatedAt:       now.Add(-30 * time.Minute),
		},
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			return videos, nil
		},
	}

	usecase := NewListVideosUsecase(videoRepo)

	output, err := usecase.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output == nil {
		t.Fatal("expected output, got nil")
	}

	if len(output.Videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(output.Videos))
	}

	video := output.Videos[0]

	if video.Status != string(entities.VideoStatusFailed) {
		t.Errorf("expected Status '%s', got '%s'", entities.VideoStatusFailed, video.Status)
	}

	if video.ErrorMessage != errorMessage {
		t.Errorf("expected ErrorMessage '%s', got '%s'", errorMessage, video.ErrorMessage)
	}
}

func TestListVideosUsecase_Execute_MultipleStatuses(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	now := time.Now()
	videos := []*entities.Video{
		{
			ID:              "video-pending",
			UserID:          userID,
			OriginalName:    "pending.mp4",
			Status:          entities.VideoStatusPending,
			ProgressPercent: 0,
			FileSize:        1024000,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "video-processing",
			UserID:          userID,
			OriginalName:    "processing.mp4",
			Status:          entities.VideoStatusProcessing,
			ProgressPercent: 45,
			FileSize:        2048000,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "video-completed",
			UserID:          userID,
			OriginalName:    "completed.mp4",
			Status:          entities.VideoStatusCompleted,
			ProgressPercent: 100,
			FileSize:        3072000,
			ProcessedS3Key:  "processed/completed.zip",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "video-failed",
			UserID:          userID,
			OriginalName:    "failed.mp4",
			Status:          entities.VideoStatusFailed,
			ProgressPercent: 25,
			FileSize:        512000,
			ErrorMessage:    "processing error",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	videoRepo := &mocks.MockVideoRepository{
		FindByUserIDFunc: func(ctx context.Context, uid string) ([]*entities.Video, error) {
			return videos, nil
		},
	}

	usecase := NewListVideosUsecase(videoRepo)

	output, err := usecase.Execute(ctx, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(output.Videos) != 4 {
		t.Fatalf("expected 4 videos, got %d", len(output.Videos))
	}

	statusCounts := make(map[string]int)
	for _, video := range output.Videos {
		statusCounts[video.Status]++
	}

	if statusCounts[string(entities.VideoStatusPending)] != 1 {
		t.Errorf("expected 1 pending video, got %d", statusCounts[string(entities.VideoStatusPending)])
	}

	if statusCounts[string(entities.VideoStatusProcessing)] != 1 {
		t.Errorf("expected 1 processing video, got %d", statusCounts[string(entities.VideoStatusProcessing)])
	}

	if statusCounts[string(entities.VideoStatusCompleted)] != 1 {
		t.Errorf("expected 1 completed video, got %d", statusCounts[string(entities.VideoStatusCompleted)])
	}

	if statusCounts[string(entities.VideoStatusFailed)] != 1 {
		t.Errorf("expected 1 failed video, got %d", statusCounts[string(entities.VideoStatusFailed)])
	}
}
