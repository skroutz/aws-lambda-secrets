locals {

  # ==============================================
  project-name        = "lambda-secrets"
  iam-resource-prefix = "SkroutzLambdaSecretsAppTest"
  # ==============================================

  ecr-name = local.project-name

  secretsmanager-path = local.project-name
}



data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
