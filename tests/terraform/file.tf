resource "local_file" "secrets_yaml_valid" {
  content = yamlencode({
    "secrets" : [
      # Full ARN
      {"name" = "${local.env-prefix}_MULTILINE", "valueFrom" = aws_secretsmanager_secret_version.multiline.arn},
      # Full ARN
      {"name" = "${local.env-prefix}_PLAIN", "valueFrom" = aws_secretsmanager_secret_version.plain.arn},
      # ARN without suffix
      {"name" = "${local.env-prefix}_JSON", "valueFrom" = "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${local.secretsmanager-path}/json"},
      # SecretsManager Path
      {"name" = "${local.env-prefix}_BINARY", "valueFrom" = "${local.secretsmanager-path}/binary"},
    ]
  })
  filename = "../application/secrets-test.yaml"
}
