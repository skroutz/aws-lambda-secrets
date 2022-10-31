# Inspired from:
# https://github.com/terraform-aws-modules/terraform-aws-lambda/blob/master/examples/container-image/main.tf

provider "docker" {
  registry_auth {
    address  = format("%v.dkr.ecr.%v.amazonaws.com", data.aws_caller_identity.current.account_id, data.aws_region.current.name)
    username = data.aws_ecr_authorization_token.token.user_name
    password = data.aws_ecr_authorization_token.token.password
  }
}

module "docker_image" {
  source = "github.com/terraform-aws-modules/terraform-aws-lambda//modules/docker-build?ref=v3.3.1"

  create_ecr_repo = true
  ecr_repo        = local.ecr-name
  ecr_repo_lifecycle_policy = jsonencode({
    "rules" : [
      {
        "rulePriority" : 1,
        "description" : "Keep only the last 2 images",
        "selection" : {
          "tagStatus" : "any",
          "countType" : "imageCountMoreThan",
          "countNumber" : 2
        },
        "action" : {
          "type" : "expire"
        }
      }
    ]
  })

  image_tag   = "latest"
  source_path = "../application"

  depends_on = [local_file.secrets_yaml_valid]
}
