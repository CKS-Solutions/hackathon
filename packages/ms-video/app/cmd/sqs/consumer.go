package sqs

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/dynamodb"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/notification"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/s3"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driven/sqs"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/usecases"
	awsinfra "github.com/cks-solutions/hackathon/ms-video/internal/infra/aws"
)

type SQSConsumer struct {
	Ctx        context.Context
	VideoQueue ports.VideoQueue
	Usecase    *usecases.ProcessVideoUsecase
}

func NewSQSConsumer(ctx context.Context, region awsinfra.Region, stage awsinfra.Stage) *SQSConsumer {
	s3Client := awsinfra.NewS3Client(region, stage)
	dynamoClient := awsinfra.NewDynamoClient(region, stage)
	sqsClient := awsinfra.NewSQSClient(region, stage)

	storageService := s3.NewS3StorageService(s3Client)
	videoRepository := dynamodb.NewDynamoVideoRepository(dynamoClient)
	videoQueue := sqs.NewSQSVideoQueue(sqsClient)
	notificationService := notification.NewNotificationService()

	processUsecase := usecases.NewProcessVideoUsecase(videoRepository, storageService, notificationService)

	return &SQSConsumer{
		Ctx:        ctx,
		VideoQueue: videoQueue,
		Usecase:    processUsecase,
	}
}

func (c *SQSConsumer) Start() {
	log.Println("🎬 Video processing worker started")

	for {
		select {
		case <-c.Ctx.Done():
			log.Println("Consumer shutting down")
			return
		default:
			messages, err := c.VideoQueue.Get(c.Ctx)
			if err != nil {
				log.Println("[QUEUE_READ] Consumer error:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(messages) == 0 {
				continue
			}

			log.Printf("Received %d messages from queue", len(messages))

			for _, message := range messages {
				var input dto.VideoProcessMessage

				data := []byte(*message.Body)

				err := json.Unmarshal(data, &input)
				if err != nil {
					log.Println("[INVALID_DATA] Consumer error:", err)
					c.VideoQueue.Delete(c.Ctx, message)
					continue
				}

				log.Printf("Processing video: %s", input.VideoID)

				err = c.Usecase.Execute(c.Ctx, input)
				if err != nil {
					log.Println("[USE_CASE_ERR] Consumer error:", err)
					continue
				}

				if err := c.VideoQueue.Delete(c.Ctx, message); err != nil {
					log.Println("[DELETE_ERR] Failed to delete message:", err)
				} else {
					log.Printf("Successfully processed and deleted message for video: %s", input.VideoID)
				}
			}
		}
	}
}
