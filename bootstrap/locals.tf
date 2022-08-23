locals {

  # ==============================================
  project-name        = "aws-lambda-secrets-layer-state"
  iam-resource-prefix = "SkroutzLambdaSecretsLayer"
  # ==============================================

  s3-bucket-name = "${local.project-name}-${data.aws_caller_identity.current.account_id}"
  dynamodb-name  = local.project-name

  iam-deployer-user   = "${local.iam-resource-prefix}DeployerUser"
  iam-deployer-role   = "${local.iam-resource-prefix}DeployerRole"
  iam-deployer-policy = "${local.iam-resource-prefix}DeployerPolicy"

  ecr-name = "aws-lambda-secrets-extension"
  lambda-layer-name = "aws-lambda-secrets-layer"

  lambda-layer-arn = "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:layer:${local.lambda-layer-name}"
  ecr-arn = "arn:aws:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/${local.ecr-name}"
}



data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
