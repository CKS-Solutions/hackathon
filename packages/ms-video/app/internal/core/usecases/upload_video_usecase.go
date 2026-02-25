package usecases

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

const (
	MaxVideoSize = 500 * 1024 * 1024 // 500MB
)

type UploadVideoUsecase struct {
	videoRepository ports.VideoRepository
	storageService  ports.StorageService
	videoQueue      ports.VideoQueue
}

func NewUploadVideoUsecase(
	videoRepository ports.VideoRepository,
	storageService ports.StorageService,
	videoQueue ports.VideoQueue,
) *UploadVideoUsecase {
	return &UploadVideoUsecase{
		videoRepository: videoRepository,
		storageService:  storageService,
		videoQueue:      videoQueue,
	}
}

func (u *UploadVideoUsecase) Execute(ctx context.Context, input dto.UploadVideoInput) (*dto.UploadVideoOutput, error) {
	if input.File.Size > MaxVideoSize {
		return nil, utils.NewBadRequestError(fmt.Sprintf("file size exceeds maximum allowed size of %dMB", MaxVideoSize/(1024*1024)))
	}

	ext := filepath.Ext(input.File.Filename)
	allowedExtensions := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".mkv":  true,
		".webm": true,
	}
	if !allowedExtensions[ext] {
		return nil, utils.NewBadRequestError("invalid video format. Allowed formats: mp4, avi, mov, mkv, webm")
	}

	file, err := input.File.Open()
	if err != nil {
		return nil, utils.NewInternalServerError("failed to open uploaded file")
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to read uploaded file")
	}

	rawS3Key := fmt.Sprintf("raw/%s/%s", input.UserID, input.File.Filename)
	video := entities.NewVideo(input.UserID, input.UserEmail, input.File.Filename, rawS3Key, input.File.Size)

	if err := u.storageService.Upload(ctx, rawS3Key, fileContent, input.File.Header.Get("Content-Type")); err != nil {
		fmt.Printf("failed to upload video to storage: %v\n", err)
		return nil, utils.NewInternalServerError("failed to upload video to storage: " + err.Error())
	}

	if err := u.videoRepository.Save(ctx, video); err != nil {
		_ = u.storageService.Delete(ctx, rawS3Key)
		return nil, utils.NewInternalServerError("failed to save video metadata")
	}

	queueMessage := dto.VideoProcessMessage{
		VideoID:   video.ID,
		UserID:    video.UserID,
		UserEmail: video.UserEmail,
		RawS3Key:  video.RawS3Key,
	}
	if err := u.videoQueue.Send(ctx, queueMessage); err != nil {
		return nil, utils.NewInternalServerError("failed to queue video for processing")
	}

	return &dto.UploadVideoOutput{
		VideoID:      video.ID,
		OriginalName: video.OriginalName,
		Status:       string(video.Status),
		Message:      "Video uploaded successfully and queued for processing",
	}, nil
}
