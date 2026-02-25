package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

type DynamoVideoRepository struct {
	client *dynamodb.Client
}

const TABLE_NAME = "MSVideo.Video"

func NewDynamoVideoRepository(client *dynamodb.Client) ports.VideoRepository {
	return &DynamoVideoRepository{
		client: client,
	}
}

func (r *DynamoVideoRepository) Save(ctx context.Context, video *entities.Video) error {
	item, err := attributevalue.MarshalMap(video)
	if err != nil {
		return fmt.Errorf("failed to marshal video: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME),
		Item:      item,
	})

	return err
}

func (r *DynamoVideoRepository) FindByID(ctx context.Context, videoID string) (*entities.Video, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: videoID},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("video not found")
	}

	var video entities.Video
	if err := attributevalue.UnmarshalMap(result.Item, &video); err != nil {
		return nil, fmt.Errorf("failed to unmarshal video: %w", err)
	}

	return &video, nil
}

func (r *DynamoVideoRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.Video, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(TABLE_NAME),
		IndexName:              aws.String("user_id-index"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	videos := make([]*entities.Video, 0, len(result.Items))
	for _, item := range result.Items {
		var video entities.Video
		if err := attributevalue.UnmarshalMap(item, &video); err != nil {
			continue
		}
		videos = append(videos, &video)
	}

	return videos, nil
}

func (r *DynamoVideoRepository) Update(ctx context.Context, video *entities.Video) error {
	item, err := attributevalue.MarshalMap(video)
	if err != nil {
		return fmt.Errorf("failed to marshal video: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME),
		Item:      item,
	})

	return err
}
