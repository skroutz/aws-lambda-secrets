locals {

  # ==============================================
  project-name        = "lambda-secrets"
  iam-resource-prefix = "SkroutzLambdaSecretsApp"
  # ==============================================

  ecr-name = local.project-name
}



data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
