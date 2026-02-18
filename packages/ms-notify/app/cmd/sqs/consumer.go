package sqs

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driven/dynamo"
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driven/ses"
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driven/sqs"
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/usecases"
	awsinfra "github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

type SQSConsumer struct {
	Ctx               context.Context
	NotificationQueue ports.NotificationQueue
	Usecase           usecases.NotificationConsumerUsecase
}

func NewSQSConsumer(ctx context.Context, region awsinfra.Region, stage awsinfra.Stage) *SQSConsumer {
	sqsClient := awsinfra.NewSQSClient(region, stage)
	dynamoClient := awsinfra.NewDynamoClient(region, stage)
	sesClient := awsinfra.NewSESClient(region, stage)

	notificationQueue := sqs.NewNotificationQueue(*sqsClient)
	notificationTable := dynamo.NewNotificationTable(*dynamoClient)
	emailService := ses.NewEmailService(*sesClient)

	usecase := usecases.NewNotificationConsumerUsecase(notificationTable, emailService)

	return &SQSConsumer{
		Ctx:               ctx,
		NotificationQueue: notificationQueue,
		Usecase:           usecase,
	}
}

func (c *SQSConsumer) Start() {
	// TODO: add graceful shutdown
	// TODO: add retry mechanism using a Dead Letter Queue (DLQ)

	for {
		select {
		case <-c.Ctx.Done():
			log.Println("Consumer shutting down")
			return
		default:
			messages, err := c.NotificationQueue.Get(c.Ctx)
			if err != nil {
				log.Println("[QUEUE_READ] Consumer error:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, message := range messages {
				input := dto.NotificationInput{}

				data := []byte(*message.Body)

				err := json.Unmarshal(data, &input)
				if err != nil {
					log.Println("[INVALID_DATA] Consumer error:", err)
					c.NotificationQueue.Delete(c.Ctx, message)
					continue
				}

				err = c.Usecase.Run(c.Ctx, input)
				if err != nil {
					log.Println("[USE_CASE_ERR] Consumer error:", err)
					continue
				}

				if err := c.NotificationQueue.Delete(c.Ctx, message); err != nil {
					log.Println("[DELETE_ERR] Failed to delete message:", err)
				}
			}
		}
	}
}
