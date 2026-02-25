package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
)

// dynamoClientInterface allows testing without a real DynamoDB client. *infra/aws.DynamoClient satisfies it.
type dynamoClientInterface interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type NotificationTableImpl struct {
	client dynamoClientInterface
}

const NOTIFICATION_TABLE_NAME = "MSNotify.Notification"

func NewNotificationTable(client dynamoClientInterface) ports.NotificationTable {
	return &NotificationTableImpl{
		client: client,
	}
}

func (t *NotificationTableImpl) Put(ctx context.Context, notification entities.NotificationDB) error {
	av, err := attributevalue.MarshalMap(notification)
	if err != nil {
		return err
	}

	_, err = t.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(NOTIFICATION_TABLE_NAME),
		Item:      av,
	})
	if err != nil {
		return err
	}

	return nil
}
