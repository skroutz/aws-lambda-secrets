# Sample IAM Policy that enables IAM User to access ECR and Create Lambdas
data "aws_iam_policy_document" "user-policy-document" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = [
      "*"
    ]
  }

  statement {
    sid = "ReadWriteECR"
    actions = [
      # Push/Pull to ECR
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:CompleteLayerUpload",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
    ]
    resources = [
      "arn:aws:ecr::${data.aws_caller_identity.current.account_id}:${local.ecr-name}"
    ]
  }
}

# Add ARN (module.user-policy.arn) to 'module.user-role.custom_role_policy_arns'
# to grant IAM Policy access to the CI/CD worker
module "user-policy" {
  source = "github.com/terraform-aws-modules/terraform-aws-iam//modules/iam-policy?ref=v5.1.0"

  name   = local.iam-deployer-policy
  policy = data.aws_iam_policy_document.user-policy-document.json
}
