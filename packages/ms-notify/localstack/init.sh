#!/bin/bash

echo "Inicializando recursos AWS no LocalStack..."

# Configurar variáveis
AWS_REGION=us-east-1
QUEUE_NAME="MSNotify-Queue"
# DLQ_NAME="MSNotify-DLQueue"
TABLE_NAME="MSNotify.Notification"
SES_EMAIL="cks.hackathon.noreply@gmail.com"

# Criar fila SQS
echo "Criando fila SQS: $QUEUE_NAME"
awslocal sqs create-queue \
  --queue-name "$QUEUE_NAME" \
  --region "$AWS_REGION"

# Criar fila SQS de Dead Letter
# echo "Criando fila SQS de Dead Letter: $DLQ_NAME"
# awslocal sqs create-queue \
#   --queue-name "$DLQ_NAME" \
#   --region "$AWS_REGION"

# Criar tabela DynamoDB
echo "Criando tabela DynamoDB: $TABLE_NAME"
awslocal dynamodb create-table \
  --table-name "$TABLE_NAME" \
  --attribute-definitions \
    AttributeName=id,AttributeType=S \
  --key-schema \
    AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$AWS_REGION"

# Verificar email no SES
echo "Verificando email no SES: $SES_EMAIL"
awslocal ses verify-email-identity \
  --email-address "$SES_EMAIL" \
  --region "$AWS_REGION"

echo "✅ Recursos AWS criados com sucesso!"
