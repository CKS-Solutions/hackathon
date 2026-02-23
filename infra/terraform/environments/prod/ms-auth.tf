locals {
  ms_auth_tags = {
    Project     = "hackathon-ms-auth"
    Managed     = "terraform"
    Environment = var.environment
    Service     = "ms-auth"
  }
}

module "ms_auth_rds" {
  source = "../../modules/rds"

  name_prefix = "${var.cluster_name}-ms-auth"
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.private_subnet_ids

  allowed_security_groups = [module.eks.cluster_security_group_id]

  db_name     = "authdb"
  db_username = "authuser"

  tags = local.ms_auth_tags
}
