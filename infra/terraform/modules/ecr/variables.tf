variable "repository_names" {
  description = "List of ECR repository names to create."
  type        = list(string)

  validation {
    condition     = length(var.repository_names) > 0
    error_message = "At least one repository name is required."
  }
}

variable "image_tag_mutability" {
  description = "Tag mutability for the repositories (MUTABLE or IMMUTABLE)."
  type        = string
  default     = "MUTABLE"
}

variable "lifecycle_policy_count" {
  description = "If set, keep only this many untagged images per repository (reduces storage). Set to null to disable."
  type        = number
  default     = null
}
