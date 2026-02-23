variable "table_name" {
  type        = string
  description = "Name of the DynamoDB table."
}

variable "hash_key" {
  type        = string
  description = "Name of the hash key attribute."
}

variable "hash_key_type" {
  type        = string
  default     = "S"
  description = "Type of the hash key (S, N, B)."
}

variable "billing_mode" {
  type        = string
  default     = "PAY_PER_REQUEST"
  description = "Billing mode: PAY_PER_REQUEST or PROVISIONED."
}

variable "tags" {
  type    = map(string)
  default = {}
}
