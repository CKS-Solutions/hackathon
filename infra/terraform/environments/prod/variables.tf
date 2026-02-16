variable "aws_region" {
  description = "AWS region for resources."
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g. dev, prod)."
  type        = string
  default     = "prod"
}

variable "availability_zones" {
  description = "List of availability zone names for VPC subnets."
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
  default     = "hackathon-prod"
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
  description = "GitHub repository in owner/repo format (for OIDC trust policy). In CI, set via TF_VAR_github_repo from github.repository."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$", var.github_repo))
    error_message = "github_repo must be in owner/repo format (e.g. myorg/my-repo)."
  }
}

variable "github_branch" {
  description = "Branch allowed to assume the GitHub Actions IAM role (e.g. master)."
  type        = string
  default     = "master"
}

variable "terraform_state_bucket" {
  description = "S3 bucket name for Terraform state (used by GitHub Actions to read/write state)."
  type        = string
}

variable "cluster_access_principal_arns" {
  description = "IAM principal ARNs (user or role) to grant EKS cluster access (kubectl). Get yours with: aws sts get-caller-identity --query Arn --output text. Set in terraform.tfvars (não commitar)."
  type        = list(string)
  default     = []
}
