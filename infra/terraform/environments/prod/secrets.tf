# Shared secrets for all services

# JWT Secret shared by ms-auth and ms-video
resource "random_password" "jwt_secret" {
  length  = 64
  special = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "aws_secretsmanager_secret" "jwt_secret" {
  name                    = "${var.cluster_name}-jwt-secret"
  description             = "JWT secret shared by ms-auth and ms-video"
  recovery_window_in_days = 7

  tags = {
    Project     = "hackathon"
    Managed     = "terraform"
    Environment = var.environment
  }
}

resource "aws_secretsmanager_secret_version" "jwt_secret" {
  secret_id = aws_secretsmanager_secret.jwt_secret.id
  secret_string = jsonencode({
    jwt_secret = random_password.jwt_secret.result
  })
}
