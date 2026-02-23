variable "queue_name" {
  type        = string
  description = "Name of the main SQS queue."
}

variable "create_dlq" {
  type        = bool
  default     = true
  description = "If true, create a dead-letter queue and attach redrive_policy to the main queue."
}

variable "dlq_name" {
  type        = string
  default     = null
  description = "Name of the DLQ. If null and create_dlq is true, derived as queue_name with '-DLQ' suffix."
}

variable "visibility_timeout_seconds" {
  type    = number
  default = 60
}

variable "message_retention_seconds" {
  type    = number
  default = 345600
}

variable "receive_wait_time_seconds" {
  type    = number
  default = 20
}

variable "max_receive_count" {
  type        = number
  default     = 5
  description = "Max receive count before message goes to DLQ (only used when create_dlq is true)."
}

variable "tags" {
  type    = map(string)
  default = {}
}
