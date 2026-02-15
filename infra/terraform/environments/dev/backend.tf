# Backend S3 for Terraform state. Create the bucket before first terraform init.
# In production, enable versioning on the bucket. For this scope we use a bucket without versioning to reduce cost.
#
# Recomendado: use um arquivo para não repetir o comando longo.
#   cp backend.hcl.example backend.hcl   # edite o bucket
#   terraform init -backend-config=backend.hcl
terraform {
  backend "s3" {}
}
