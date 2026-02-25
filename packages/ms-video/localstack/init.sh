#!/bin/bash

echo "Initializing LocalStack resources for ms-video..."

# Wait for LocalStack to be ready
sleep 5

# Create S3 bucket
awslocal s3 mb s3://video-system
echo "✓ Created S3 bucket: video-system"

# Create SQS queue
awslocal sqs create-queue --queue-name MSVideo-Queue
echo "✓ Created SQS queue: MSVideo-Queue"

# Create DynamoDB table
awslocal dynamodb create-table \
    --table-name MSVideo.Video \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"user_id-index\",
                \"KeySchema\": [
                    {\"AttributeName\":\"user_id\",\"KeyType\":\"HASH\"},
                    {\"AttributeName\":\"created_at\",\"KeyType\":\"RANGE\"}
                ],
                \"Projection\": {
                    \"ProjectionType\":\"ALL\"
                },
                \"ProvisionedThroughput\": {
                    \"ReadCapacityUnits\": 5,
                    \"WriteCapacityUnits\": 5
                }
            }
        ]" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5

echo "✓ Created DynamoDB table: MSVideo.Video with user_id-index"

echo "LocalStack initialization complete!"
