package usecases

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

type ProcessVideoUsecase struct {
	videoRepository     ports.VideoRepository
	storageService      ports.StorageService
	notificationService ports.NotificationService
}

func NewProcessVideoUsecase(
	videoRepository ports.VideoRepository,
	storageService ports.StorageService,
	notificationService ports.NotificationService,
) *ProcessVideoUsecase {
	return &ProcessVideoUsecase{
		videoRepository:     videoRepository,
		storageService:      storageService,
		notificationService: notificationService,
	}
}

func (u *ProcessVideoUsecase) Execute(ctx context.Context, message dto.VideoProcessMessage) error {
	video, err := u.videoRepository.FindByID(ctx, message.VideoID)
	if err != nil {
		log.Printf("Failed to find video %s: %v", message.VideoID, err)
		return err
	}

	video.UpdateProgress(10, entities.VideoStatusProcessing)
	if err := u.videoRepository.Update(ctx, video); err != nil {
		log.Printf("Failed to update video status: %v", err)
		return err
	}

	log.Printf("Downloading video from S3: %s", message.RawS3Key)
	videoData, err := u.storageService.Download(ctx, message.RawS3Key)
	if err != nil {
		video.MarkAsFailed(fmt.Sprintf("failed to download raw video: %v", err))
		u.videoRepository.Update(ctx, video)
		
		if message.UserEmail != "" {
			notifyErr := u.notificationService.SendVideoFailedNotification(ctx, message.UserEmail, video.ID, video.OriginalName, fmt.Sprintf("Failed to download raw video: %v", err))
			if notifyErr != nil {
				log.Printf("Failed to send failure notification: %v", notifyErr)
			}
		}
		
		return err
	}

	video.UpdateProgress(30, entities.VideoStatusProcessing)
	u.videoRepository.Update(ctx, video)

	log.Printf("Processing video %s", message.VideoID)
	
	time.Sleep(2 * time.Second)
	
	video.UpdateProgress(60, entities.VideoStatusProcessing)
	u.videoRepository.Update(ctx, video)

	log.Printf("Creating ZIP file for video %s", message.VideoID)
	zipData, err := u.createZipFile(video.OriginalName, videoData)
	if err != nil {
		video.MarkAsFailed(fmt.Sprintf("failed to create zip file: %v", err))
		u.videoRepository.Update(ctx, video)
		
		if message.UserEmail != "" {
			notifyErr := u.notificationService.SendVideoFailedNotification(ctx, message.UserEmail, video.ID, video.OriginalName, fmt.Sprintf("Failed to create zip file: %v", err))
			if notifyErr != nil {
				log.Printf("Failed to send failure notification: %v", notifyErr)
			}
		}
		
		return err
	}
		
	video.UpdateProgress(80, entities.VideoStatusProcessing)
	u.videoRepository.Update(ctx, video)

	processedS3Key := fmt.Sprintf("processed/%s/%s.zip", message.UserID, video.ID)
	log.Printf("Uploading processed video to S3: %s", processedS3Key)
	if err := u.storageService.Upload(ctx, processedS3Key, zipData, "application/zip"); err != nil {
		video.MarkAsFailed(fmt.Sprintf("failed to upload processed video: %v", err))
		u.videoRepository.Update(ctx, video)

		if message.UserEmail != "" {
			notifyErr := u.notificationService.SendVideoFailedNotification(ctx, message.UserEmail, video.ID, video.OriginalName, fmt.Sprintf("Failed to upload processed video: %v", err))
			if notifyErr != nil {
				log.Printf("Failed to send failure notification: %v", notifyErr)
			}
		}

		return err
	}

	log.Printf("Video processing completed: %s", message.VideoID)
	video.MarkAsCompleted(processedS3Key)
	if err := u.videoRepository.Update(ctx, video); err != nil {
		log.Printf("Failed to mark video as completed: %v", err)
		return err
	}

	if message.UserEmail != "" {
		notifyErr := u.notificationService.SendVideoProcessedNotification(ctx, message.UserEmail, video.ID, video.OriginalName)
		if notifyErr != nil {
			log.Printf("Failed to send success notification: %v", notifyErr)
		}
	}

	return nil
}

func (u *ProcessVideoUsecase) createZipFile(originalName string, videoData []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	videoFile, err := zipWriter.Create(originalName)
	if err != nil {
		return nil, err
	}
	if _, err := videoFile.Write(videoData); err != nil {
		return nil, err
	}

	metadataFile, err := zipWriter.Create("metadata.txt")
	if err != nil {
		return nil, err
	}
	metadata := fmt.Sprintf("Original File: %s\nProcessed: %s\nSize: %d bytes\n", 
		originalName, 
		time.Now().Format(time.RFC3339), 
		len(videoData))
	if _, err := metadataFile.Write([]byte(metadata)); err != nil {
		return nil, err
	}

	readmeFile, err := zipWriter.Create("README.txt")
	if err != nil {
		return nil, err
	}
	readme := fmt.Sprintf("Video Processing Complete\n\nOriginal file: %s\nProcessed on: %s\n\nThis archive contains your processed video file.\n",
		originalName,
		time.Now().Format("2006-01-02 15:04:05"))
	if _, err := readmeFile.Write([]byte(readme)); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
