variable "name_prefix" {
  description = "Prefix for resource names (e.g., environment or project name)."
  type        = string
}

variable "identifier" {
  description = "Identifier for the RDS instance. If not provided, defaults to {name_prefix}-db."
  type        = string
  default     = null
}

variable "vpc_id" {
  description = "VPC ID where the RDS instance will be created."
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the DB subnet group (should be private subnets)."
  type        = list(string)
}

variable "allowed_security_groups" {
  description = "List of security group IDs allowed to access the RDS instance."
  type        = list(string)
  default     = []
}

variable "allowed_cidr_blocks" {
  description = "List of CIDR blocks allowed to access the RDS instance."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "engine" {
  description = "Database engine (e.g., postgres, mysql, mariadb)."
  type        = string
  default     = "postgres"
}

variable "engine_version" {
  description = "Database engine version."
  type        = string
  default     = "15.4"
}

variable "instance_class" {
  description = "RDS instance class (e.g., db.t3.micro, db.t3.small)."
  type        = string
  default     = "db.t3.micro"
}

variable "allocated_storage" {
  description = "Initial allocated storage in GB."
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Maximum allocated storage for autoscaling (0 to disable)."
  type        = number
  default     = 20
}

variable "storage_type" {
  description = "Storage type (gp2, gp3, io1, io2)."
  type        = string
  default     = "gp2"
}

variable "storage_encrypted" {
  description = "Enable storage encryption."
  type        = bool
  default     = false
}

variable "kms_key_id" {
  description = "KMS key ID for storage encryption. If not provided, uses the default RDS KMS key."
  type        = string
  default     = null
}

# Database Configuration
variable "db_name" {
  description = "Name of the database to create."
  type        = string
  default     = "authdb"
}

variable "db_username" {
  description = "Master username for the database."
  type        = string
  default     = "dbadmin"
}

variable "db_password" {
  description = "Master password for the database. If not provided, a random password will be generated."
  type        = string
  default     = null
  sensitive   = true
}

variable "port" {
  description = "Database port."
  type        = number
  default     = 5432
}

variable "publicly_accessible" {
  description = "Whether the DB instance is publicly accessible."
  type        = bool
  default     = true
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment for high availability."
  type        = bool
  default     = false
}

variable "backup_retention_period" {
  description = "Number of days to retain backups (0 to disable)."
  type        = number
  default     = 1
}

variable "backup_window" {
  description = "Preferred backup window (UTC)."
  type        = string
  default     = "03:00-04:00"
}

variable "maintenance_window" {
  description = "Preferred maintenance window (UTC)."
  type        = string
  default     = "sun:04:00-sun:05:00"
}

variable "delete_automated_backups" {
  description = "Whether to delete automated backups immediately when the DB instance is deleted."
  type        = bool
  default     = true
}

variable "skip_final_snapshot" {
  description = "Whether to skip final snapshot when deleting the instance."
  type        = bool
  default     = true
}

variable "enabled_cloudwatch_logs_exports" {
  description = "List of log types to export to CloudWatch (e.g., ['postgresql', 'upgrade'])."
  type        = list(string)
  default     = []
}

variable "monitoring_interval" {
  description = "Enhanced monitoring interval in seconds (0, 1, 5, 10, 15, 30, 60)."
  type        = number
  default     = 0
}

variable "monitoring_role_arn" {
  description = "ARN of IAM role for enhanced monitoring. Required if monitoring_interval > 0."
  type        = string
  default     = null
}

variable "performance_insights_enabled" {
  description = "Enable Performance Insights."
  type        = bool
  default     = false
}

variable "performance_insights_retention_period" {
  description = "Retention period for Performance Insights data in days (7 or 731)."
  type        = number
  default     = 7
}

variable "auto_minor_version_upgrade" {
  description = "Enable automatic minor version upgrades."
  type        = bool
  default     = true
}

variable "deletion_protection" {
  description = "Enable deletion protection."
  type        = bool
  default     = false
}

variable "apply_immediately" {
  description = "Apply changes immediately instead of during maintenance window."
  type        = bool
  default     = false
}

variable "parameter_group_name" {
  description = "Name of the DB parameter group to associate with this instance."
  type        = string
  default     = null
}

variable "create_secret" {
  description = "Whether to create a secret in AWS Secrets Manager with database credentials."
  type        = bool
  default     = true
}

variable "secret_recovery_window_in_days" {
  description = "Number of days to retain secret after deletion (0 to force delete)."
  type        = number
  default     = 0
}

variable "tags" {
  description = "Tags to apply to all resources."
  type        = map(string)
  default     = {}
}
