provider "aws" {
  region = var.aws_region
}

module "vpc" {
  source = "../../modules/vpc"

  name_prefix         = var.cluster_name
  vpc_cidr            = var.vpc_cidr
  availability_zones  = var.availability_zones
  tags = {
    Environment = var.environment
  }
}

module "eks" {
  source = "../../modules/eks"

  cluster_name         = var.cluster_name
  cluster_version      = var.cluster_version
  vpc_id               = module.vpc.vpc_id
  private_subnet_ids   = module.vpc.private_subnet_ids
  node_instance_types = ["t3.small"]
  node_desired_size    = 2
  node_min_size        = 1
  node_max_size        = 3
  environment          = var.environment
}
