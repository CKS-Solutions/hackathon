package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-video/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

type SQSVideoQueue struct {
	client *sqs.Client
}

const QUEUE_NAME = "MSVideo-Queue"

func NewSQSVideoQueue(client *sqs.Client) ports.VideoQueue {
	return &SQSVideoQueue{
		client: client,
	}
}

func (i *SQSVideoQueue) getQueueUrl() (*string, error) {
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

func (q *SQSVideoQueue) Send(ctx context.Context, message dto.VideoProcessMessage) error {
	url, err := q.getQueueUrl()
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    url,
		MessageBody: aws.String(string(messageBody)),
	})

	return err
}

func (q *SQSVideoQueue) Get(ctx context.Context) ([]sqstypes.Message, error) {
	url, err := q.getQueueUrl()
	if err != nil {
		return nil, err
	}

	result, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            url,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   300,
	})

	if err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func (q *SQSVideoQueue) Delete(ctx context.Context, message sqstypes.Message) error {
	url, err := q.getQueueUrl()
	if err != nil {
		return err
	}

	_, err = q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      url,
		ReceiptHandle: message.ReceiptHandle,
	})

	return err
}
