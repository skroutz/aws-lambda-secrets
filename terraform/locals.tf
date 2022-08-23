locals {

  # ==============================================
  project-name        = "aws-lambda-secrets-extension"
  iam-resource-prefix = "SkroutzLambdaSecretsApp"
  # ==============================================

  ecr-name = local.project-name
}



data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
