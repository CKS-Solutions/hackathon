package sqs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
)

const QUEUE_NAME = "MSNotify-Queue"

// sqsClientInterface allows testing without a real SQS client. *infra/aws.SQSClient satisfies it.
type sqsClientInterface interface {
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type NotificationQueueImpl struct {
	client sqsClientInterface
}

func NewNotificationQueue(client sqsClientInterface) ports.NotificationQueue {
	return &NotificationQueueImpl{
		client: client,
	}
}

func (i *NotificationQueueImpl) getQueueUrl() (*string, error) {
	if url := os.Getenv("SQS_QUEUE_URL"); url != "" {
		return &url, nil
	}
	queueUrl, err := i.client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(QUEUE_NAME),
	})
	if err != nil {
		return aws.String(""), err
	}

	return queueUrl.QueueUrl, nil
}

func (i *NotificationQueueImpl) Push(ctx context.Context, message entities.Notification) error {
	url, err := i.getQueueUrl()
	if err != nil {
		return err
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = i.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    url,
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *NotificationQueueImpl) Get(ctx context.Context) ([]types.Message, error) {
	notifications := []types.Message{}

	url, err := i.getQueueUrl()
	if err != nil {
		return notifications, err
	}

	out, err := i.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            url,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   30,
	})
	if err != nil {
		return notifications, err
	}

	return out.Messages, nil
}

func (i *NotificationQueueImpl) Delete(ctx context.Context, message types.Message) error {
	url, err := i.getQueueUrl()
	if err != nil {
		return err
	}

	_, err = i.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      url,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		return err
	}

	return nil
}
