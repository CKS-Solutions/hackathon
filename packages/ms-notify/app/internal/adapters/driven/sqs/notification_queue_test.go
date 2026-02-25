package sqs

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type fakeSQSClient struct {
	getQueueUrlErr    error
	sendMessageErr    error
	receiveMessageOut *sqs.ReceiveMessageOutput
	receiveMessageErr error
	deleteMessageErr  error
}

func (f *fakeSQSClient) GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	if f.getQueueUrlErr != nil {
		return nil, f.getQueueUrlErr
	}
	url := "https://sqs.us-east-1.amazonaws.com/123/queue"
	return &sqs.GetQueueUrlOutput{QueueUrl: &url}, nil
}

func (f *fakeSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if f.sendMessageErr != nil {
		return nil, f.sendMessageErr
	}
	return &sqs.SendMessageOutput{}, nil
}

func (f *fakeSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	if f.receiveMessageErr != nil {
		return nil, f.receiveMessageErr
	}
	if f.receiveMessageOut != nil {
		return f.receiveMessageOut, nil
	}
	return &sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil
}

func (f *fakeSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	if f.deleteMessageErr != nil {
		return nil, f.deleteMessageErr
	}
	return &sqs.DeleteMessageOutput{}, nil
}

func TestNotificationQueueImpl_Push(t *testing.T) {
	ctx := context.Background()
	msg := entities.Notification{Id: "1", Subject: "s", From: "f@x.com", To: []string{"t@x.com"}, Html: "h"}

	t.Run("success", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{}).(*NotificationQueueImpl)
		err := queue.Push(ctx, msg)
		if err != nil {
			t.Errorf("Push: %v", err)
		}
	})

	t.Run("getQueueUrl error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{getQueueUrlErr: errors.New("no queue")}).(*NotificationQueueImpl)
		err := queue.Push(ctx, msg)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("SendMessage error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{sendMessageErr: errors.New("send failed")}).(*NotificationQueueImpl)
		err := queue.Push(ctx, msg)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestNotificationQueueImpl_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("success empty", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{}).(*NotificationQueueImpl)
		msgs, err := queue.Get(ctx)
		if err != nil {
			t.Errorf("Get: %v", err)
		}
		if len(msgs) != 0 {
			t.Errorf("len(msgs) = %d", len(msgs))
		}
	})

	t.Run("getQueueUrl error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{getQueueUrlErr: errors.New("no queue")}).(*NotificationQueueImpl)
		msgs, err := queue.Get(ctx)
		if err == nil {
			t.Error("expected error")
		}
		if len(msgs) != 0 {
			t.Errorf("expected empty slice, got len %d", len(msgs))
		}
	})

	t.Run("ReceiveMessage error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{receiveMessageErr: errors.New("receive failed")}).(*NotificationQueueImpl)
		msgs, err := queue.Get(ctx)
		if err == nil {
			t.Error("expected error")
		}
		if msgs == nil {
			t.Error("expected non-nil slice")
		}
	})
}

func TestNotificationQueueImpl_Delete(t *testing.T) {
	ctx := context.Background()
	message := types.Message{ReceiptHandle: aws.String("handle"), Body: aws.String("{}")}

	t.Run("success", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{}).(*NotificationQueueImpl)
		err := queue.Delete(ctx, message)
		if err != nil {
			t.Errorf("Delete: %v", err)
		}
	})

	t.Run("getQueueUrl error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{getQueueUrlErr: errors.New("no queue")}).(*NotificationQueueImpl)
		err := queue.Delete(ctx, message)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("DeleteMessage error", func(t *testing.T) {
		queue := NewNotificationQueue(&fakeSQSClient{deleteMessageErr: errors.New("delete failed")}).(*NotificationQueueImpl)
		err := queue.Delete(ctx, message)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestNotificationQueueImpl_getQueueUrl_env(t *testing.T) {
	const key = "SQS_QUEUE_URL"
	restore := os.Getenv(key)
	defer os.Setenv(key, restore)

	os.Setenv(key, "https://env-queue-url")
	queue := NewNotificationQueue(&fakeSQSClient{getQueueUrlErr: errors.New("should not be called")}).(*NotificationQueueImpl)
	url, err := queue.getQueueUrl()
	if err != nil {
		t.Fatalf("getQueueUrl: %v", err)
	}
	if url == nil || *url != "https://env-queue-url" {
		t.Errorf("url = %v", url)
	}
}
