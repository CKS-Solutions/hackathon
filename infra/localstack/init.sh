#!/bin/bash

AWS_REGION=us-east-1

echo "Initializing LocalStack resources for ms-video..."

# Wait for LocalStack to be ready
sleep 5

MSVIDEO_BUCKET_NAME="cks-hackathon-video-system"
MSVIDEO_QUEUE_NAME="MSVideo-Queue"
MSVIDEO_TABLE_NAME="MSVideo.Video"

# Create S3 bucket
awslocal s3 mb s3://$MSVIDEO_BUCKET_NAME
echo "✓ Created S3 bucket: $MSVIDEO_BUCKET_NAME"

# Create SQS queue
awslocal sqs create-queue --queue-name "$MSVIDEO_QUEUE_NAME" --region "$AWS_REGION"
echo "✓ Created SQS queue: $MSVIDEO_QUEUE_NAME"

# Create DynamoDB table
awslocal dynamodb create-table \
    --table-name "$MSVIDEO_TABLE_NAME" \
    --region "$AWS_REGION" \
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

echo "✓ Created DynamoDB table: $MSVIDEO_TABLE_NAME with user_id-index"

echo "Initializing LocalStack resources for ms-notify..."

MSNOTIFY_QUEUE_NAME="MSNotify-Queue"
MSNOTIFY_TABLE_NAME="MSNotify.Notification"
SES_EMAIL="cks.hackathon.noreply@gmail.com"

# Criar fila SQS
awslocal sqs create-queue \
  --queue-name "$MSNOTIFY_QUEUE_NAME" \
  --region "$AWS_REGION"

echo "✓ Created SQS queue: $MSNOTIFY_QUEUE_NAME"

# Criar tabela DynamoDB
awslocal dynamodb create-table \
  --table-name "$MSNOTIFY_TABLE_NAME" \
  --attribute-definitions \
    AttributeName=id,AttributeType=S \
  --key-schema \
    AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$AWS_REGION"

echo "✓ Created DynamoDB table: $MSNOTIFY_TABLE_NAME"

# Verificar email no SES
awslocal ses verify-email-identity \
  --email-address "$SES_EMAIL" \
  --region "$AWS_REGION"

echo "✓ Verified email in SES: $SES_EMAIL"

echo "LocalStack initialization complete!"
