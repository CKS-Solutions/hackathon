variable "name_prefix" {
  description = "Prefix for VPC and resource names (e.g. environment or project name)."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zone names for subnets."
  type        = list(string)
}

variable "tags" {
  description = "Tags to apply to VPC and related resources."
  type        = map(string)
  default     = {}
}

variable "cluster_name" {
  description = "EKS cluster name; when set, public/private subnets get tags for AWS Load Balancer Controller discovery."
  type        = string
  default     = null
}
