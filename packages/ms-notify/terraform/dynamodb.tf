resource "aws_dynamodb_table" "main" {
  name         = "MSNotify.Notification"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}
