locals {
  az_count = length(var.availability_zones)
  # Derive private/public subnet CIDRs from vpc_cidr (e.g. 10.0.0.0/16 -> 10.0.1.0/24, 10.0.2.0/24 ... for private; 10.0.101.0/24 ... for public)
  private_subnets = [for i in range(local.az_count) : cidrsubnet(var.vpc_cidr, 8, i + 1)]
  public_subnets  = [for i in range(local.az_count) : cidrsubnet(var.vpc_cidr, 8, i + 101)]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = var.name_prefix
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = local.private_subnets
  public_subnets  = local.public_subnets

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = var.tags

  # AWS Load Balancer Controller discovers subnets by these tags (Entrega 6).
  public_subnet_tags = var.cluster_name != null ? {
    "kubernetes.io/role/elb"                    = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  } : {}
  private_subnet_tags = var.cluster_name != null ? {
    "kubernetes.io/role/internal-elb"           = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  } : {}
}
