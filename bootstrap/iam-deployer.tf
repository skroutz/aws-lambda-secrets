data "aws_iam_policy_document" "user-trust-policy" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
      # Github Actions tag the Sessions with specific metadata
      "sts:TagSession"
    ]

    principals {
      type = "AWS"
      identifiers = [
        # Only the created user (deployer) can assume the Role
        module.user.iam_user_arn
      ]
    }
  }
}

module "user-role" {
  source = "github.com/terraform-aws-modules/terraform-aws-iam//modules/iam-assumable-role?ref=v5.1.0"

  create_role = true

  role_name         = local.iam-deployer-role
  role_requires_mfa = false

  custom_role_trust_policy = data.aws_iam_policy_document.user-trust-policy.json
  custom_role_policy_arns = [

    # Grants access to all Terraform State AWS Resources (S3, KMS, Dynamo)
    module.remote_state.terraform_iam_policy.arn,

    # Grants access to ECR related Actions
    module.user-policy.arn,
  ]
}

# Deployer User that will be bound to Github Workflows
module "user" {
  source = "github.com/terraform-aws-modules/terraform-aws-iam//modules/iam-user?ref=v5.1.0"

  create_iam_access_key = true
  create_user           = true
  name                  = local.iam-deployer-user
}
