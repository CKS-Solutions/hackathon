package usecases

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"

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

	log.Printf("Extracting frames from video %s", message.VideoID)
	frames, err := u.extractFrames(videoData, video.ID, video.OriginalName)
	if err != nil {
		video.MarkAsFailed(fmt.Sprintf("failed to extract frames: %v", err))
		u.videoRepository.Update(ctx, video)
		
		if message.UserEmail != "" {
			notifyErr := u.notificationService.SendVideoFailedNotification(ctx, message.UserEmail, video.ID, video.OriginalName, fmt.Sprintf("Failed to extract frames: %v", err))
			if notifyErr != nil {
				log.Printf("Failed to send failure notification: %v", notifyErr)
			}
		}
		
		return err
	}
	
	video.UpdateProgress(60, entities.VideoStatusProcessing)
	u.videoRepository.Update(ctx, video)

	log.Printf("Creating ZIP file for video %s with %d frames", message.VideoID, len(frames))
	zipData, err := u.createZipFile(video.OriginalName, videoData, frames)
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

func (u *ProcessVideoUsecase) extractFrames(videoData []byte, videoID, originalName string) ([][]byte, error) {
	tmpDir, err := ioutil.TempDir("", "video-processing-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoPath := filepath.Join(tmpDir, "input"+filepath.Ext(originalName))
	if err := ioutil.WriteFile(videoPath, videoData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write video file: %w", err)
	}

	framesDir := filepath.Join(tmpDir, "frames")
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create frames dir: %w", err)
	}

	log.Printf("Extracting frames from %s", videoPath)
	outputPattern := filepath.Join(framesDir, "frame_%04d.jpg")
	
	err = ffmpeg.Input(videoPath).Filter("fps", ffmpeg.Args{"1"}).Output(outputPattern, ffmpeg.KwArgs{
		"q:v": "2",
	}).OverWriteOutput().ErrorToStdOut().Run()
	
	if err != nil {
		return nil, fmt.Errorf("failed to extract frames with ffmpeg: %w", err)
	}

	files, err := ioutil.ReadDir(framesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read frames directory: %w", err)
	}

	var frames [][]byte
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		framePath := filepath.Join(framesDir, file.Name())
		frameData, err := ioutil.ReadFile(framePath)
		if err != nil {
			log.Printf("Failed to read frame %s: %v", file.Name(), err)
			continue
		}
		frames = append(frames, frameData)
	}

	log.Printf("Extracted %d frames from video %s", len(frames), videoID)
	return frames, nil
}

func (u *ProcessVideoUsecase) createZipFile(originalName string, videoData []byte, frames [][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	videoFile, err := zipWriter.Create(originalName)
	if err != nil {
		return nil, err
	}
	if _, err := videoFile.Write(videoData); err != nil {
		return nil, err
	}

	for i, frameData := range frames {
		frameName := fmt.Sprintf("frames/frame_%04d.jpg", i+1)
		frameFile, err := zipWriter.Create(frameName)
		if err != nil {
			return nil, fmt.Errorf("failed to create frame file in zip: %w", err)
		}
		if _, err := frameFile.Write(frameData); err != nil {
			return nil, fmt.Errorf("failed to write frame data: %w", err)
		}
	}

	metadataFile, err := zipWriter.Create("metadata.txt")
	if err != nil {
		return nil, err
	}
	metadata := fmt.Sprintf("Original File: %s\nProcessed: %s\nSize: %d bytes\nFrames Extracted: %d\n", 
		originalName, 
		time.Now().Format(time.RFC3339), 
		len(videoData),
		len(frames))
	if _, err := metadataFile.Write([]byte(metadata)); err != nil {
		return nil, err
	}

	readmeFile, err := zipWriter.Create("README.txt")
	if err != nil {
		return nil, err
	}
	readme := fmt.Sprintf("Video Processing Complete\n\nOriginal file: %s\nProcessed on: %s\nFrames extracted: %d frames\n\nThis archive contains:\n- Original video file\n- Extracted frames (1 frame per second) in the 'frames' folder\n",
		originalName,
		time.Now().Format("2006-01-02 15:04:05"),
		len(frames))
	if _, err := readmeFile.Write([]byte(readme)); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
