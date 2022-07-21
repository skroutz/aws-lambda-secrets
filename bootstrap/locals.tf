locals {

  # ==============================================
  project-name        = "lambda-secrets-app-state"
  iam-resource-prefix = "SkroutzLambdaSecretsApp"
  # ==============================================

  s3-bucket-name = "${local.project-name}-${data.aws_caller_identity.current.account_id}"
  dynamodb-name  = local.project-name

  iam-deployer-user   = "${local.iam-resource-prefix}DeployerUser"
  iam-deployer-role   = "${local.iam-resource-prefix}DeployerRole"
  iam-deployer-policy = "${local.iam-resource-prefix}DeployerPolicy"

  ecr-name = "lambda-secrets"
}



data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
