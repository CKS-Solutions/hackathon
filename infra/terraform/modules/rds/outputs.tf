output "db_instance_id" {
  description = "ID of the RDS instance."
  value       = aws_db_instance.main.id
}

output "db_instance_arn" {
  description = "ARN of the RDS instance."
  value       = aws_db_instance.main.arn
}

output "db_instance_endpoint" {
  description = "Connection endpoint for the RDS instance."
  value       = aws_db_instance.main.endpoint
}

output "db_instance_address" {
  description = "Hostname of the RDS instance."
  value       = aws_db_instance.main.address
}

output "db_instance_port" {
  description = "Port of the RDS instance."
  value       = aws_db_instance.main.port
}

output "db_name" {
  description = "Name of the database."
  value       = aws_db_instance.main.db_name
}

output "db_username" {
  description = "Master username for the database."
  value       = aws_db_instance.main.username
  sensitive   = true
}

output "db_password" {
  description = "Master password for the database."
  value       = local.db_password
  sensitive   = true
}

output "security_group_id" {
  description = "ID of the RDS security group."
  value       = aws_security_group.rds.id
}

output "db_subnet_group_name" {
  description = "Name of the DB subnet group."
  value       = aws_db_subnet_group.main.name
}

output "db_subnet_group_arn" {
  description = "ARN of the DB subnet group."
  value       = aws_db_subnet_group.main.arn
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret containing database credentials."
  value       = var.create_secret ? aws_secretsmanager_secret.db_password[0].arn : null
}

output "secret_name" {
  description = "Name of the Secrets Manager secret containing database credentials."
  value       = var.create_secret ? aws_secretsmanager_secret.db_password[0].name : null
}

output "connection_string" {
  description = "PostgreSQL connection string (for reference only, do not expose in logs)."
  value       = "${var.engine}://${var.db_username}:${local.db_password}@${aws_db_instance.main.endpoint}/${var.db_name}"
  sensitive   = true
}
