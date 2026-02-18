package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
	awsinfra "github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

type NotificationTableImpl struct {
	client awsinfra.DynamoClient
}

const NOTIFICATION_TABLE_NAME = "MSNotify.Notification"

func NewNotificationTable(client awsinfra.DynamoClient) ports.NotificationTable {
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
