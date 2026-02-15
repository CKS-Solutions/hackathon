variable "aws_region" {
  description = "AWS region for resources."
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g. dev, prod)."
  type        = string
  default     = "dev"
}

variable "availability_zones" {
  description = "List of availability zone names for VPC subnets."
  type        = list(string)
}

variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "cluster_version" {
  description = "Kubernetes version for the EKS cluster."
  type        = string
  default     = "1.31"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "github_repo" {
  description = "GitHub repository in owner/repo format (for OIDC trust policy)."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$", var.github_repo))
    error_message = "github_repo must be in owner/repo format (e.g. myorg/my-repo)."
  }
}

variable "github_branch" {
  description = "Branch allowed to assume the GitHub Actions IAM role (e.g. main)."
  type        = string
  default     = "main"
}
