package sqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
	aws_infra "github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

const QUEUE_NAME = "MSNotify-Queue"

type NotificationQueueImpl struct {
	client aws_infra.SQSClient
}

func NewNotificationQueue(client aws_infra.SQSClient) ports.NotificationQueue {
	return &NotificationQueueImpl{
		client: client,
	}
}

func (i *NotificationQueueImpl) getQueueUrl() (*string, error) {
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
