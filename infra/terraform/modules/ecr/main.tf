resource "aws_ecr_repository" "this" {
  for_each = toset(var.repository_names)

  name                 = each.key
  image_tag_mutability = var.image_tag_mutability
}

resource "aws_ecr_lifecycle_policy" "this" {
  for_each = coalesce(var.lifecycle_policy_count, 0) > 0 ? toset(var.repository_names) : toset([])

  repository = aws_ecr_repository.this[each.key].name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last ${var.lifecycle_policy_count} images"
        selection = {
          tagStatus   = "untagged"
          countType   = "imageCountMoreThan"
          countNumber = var.lifecycle_policy_count
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
