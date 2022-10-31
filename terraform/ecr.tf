module "ecr" {
  source = "github.com/terraform-aws-modules/terraform-aws-ecr?ref=v1.1.1"

  repository_name = local.project-name

  repository_image_tag_mutability = "MUTABLE"

  # Sample Lifecyle Rule to avoid clutter
  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep last 10 images",
        selection = {
          tagStatus     = "tagged",
          tagPrefixList = ["v"],
          countType     = "imageCountMoreThan",
          countNumber   = 10
        },
        action = {
          type = "expire"
        }
      }
    ]
  })
}
