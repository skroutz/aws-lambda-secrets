data "aws_iam_policy_document" "lambda-trust-policy" {
    statement {
        principals {
            type = "Service"
            identifiers = ["lambda.amazonaws.com"]
        }
        # Hardening: make sure only the dedicated Lambda can assume this role
        actions = [
            "sts:AssumeRole"
        ]
    }
}

data "aws_iam_policy_document" "lambda-policy" {
    statement {
        sid = "ReadSecrets"
        effect = "Allow"

        actions = [
            "secretsmanager:GetSecretValue",
            "kms:Decrypt"
        ]

        resources = [
            # All secrets under "lambda-secrets/" path are Readable with this Permission
            "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${local.secretsmanager-path}/*",
            "arn:aws:kms::${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/aws/secretsmanager"
        ]
    }
}

module "lambda_policy" {
    source = "github.com/terraform-aws-modules/terraform-aws-iam//modules/iam-policy?ref=v5.1.0"

    name = "${local.iam-execution-policy}"
    description = ""
    policy = data.aws_iam_policy_document.lambda-policy.json

    tags = local.tags
}

# Role to be assumed by Lambda Runtime
module "lambda_role" {
    source = "github.com/terraform-aws-modules/terraform-aws-iam//modules/iam-assumable-role?ref=v5.1.0"

    create_role = true

    role_name = local.iam-execution-role
    role_requires_mfa = false

    custom_role_trust_policy = data.aws_iam_policy_document.lambda-trust-policy.json
    custom_role_policy_arns = [
        "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
        module.lambda_policy.arn,
    ]

    tags = local.tags

}
