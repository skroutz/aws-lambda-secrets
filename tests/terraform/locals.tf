locals {

  # ==============================================
  project-name        = "lambda-secrets-test"
  iam-resource-prefix = "SkroutzLambdaSecretsAppTest"
  # ==============================================

  env-prefix = "LAMBDASECRETS"

  ecr-name = local.project-name

  secretsmanager-path = local.project-name

  lambda-container-name = "${local.project-name}-container"

  iam-execution-role   = "${local.iam-resource-prefix}ExecutionRole"
  iam-execution-policy = "${local.iam-resource-prefix}ExecutionPolicy"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_ecr_authorization_token" "token" {}
