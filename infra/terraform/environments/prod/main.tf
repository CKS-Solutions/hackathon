provider "aws" {
  region = var.aws_region
}

module "ecr" {
  source = "../../modules/ecr"

  repository_names = ["ms-auth", "ms-video", "ms-notify"]
}

# For "apply rápido" (só Entrega 4), comment out the vpc and eks modules below, then uncomment for Entrega 5.
module "vpc" {
  source = "../../modules/vpc"

  name_prefix        = var.cluster_name
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  tags = {
    Environment = var.environment
  }
}

# module "eks" {
#   source = "../../modules/eks"

#   cluster_name        = var.cluster_name
#   cluster_version     = var.cluster_version
#   vpc_id              = module.vpc.vpc_id
#   private_subnet_ids  = module.vpc.private_subnet_ids
#   node_instance_types = ["t3.small"]
#   node_desired_size   = 2
#   node_min_size       = 1
#   node_max_size       = 3
#   environment         = var.environment
# }

# EKS Access Entries: quem pode usar kubectl (sem isso dá "server asked for credentials")
# resource "aws_eks_access_entry" "cluster_access" {
#   for_each = toset(var.cluster_access_principal_arns)

#   cluster_name  = module.eks.cluster_id
#   principal_arn = each.value
#   type          = "STANDARD"
# }

# resource "aws_eks_access_policy_association" "cluster_admin" {
#   for_each = toset(var.cluster_access_principal_arns)

#   cluster_name  = module.eks.cluster_id
#   principal_arn = each.value
#   policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
#   access_scope {
#     type = "cluster"
#   }

#   depends_on = [aws_eks_access_entry.cluster_access]
# }
