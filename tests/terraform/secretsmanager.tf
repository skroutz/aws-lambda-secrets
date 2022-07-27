resource "aws_secretsmanager_secret" "plain" {
  name = "${local.secretsmanager-path}/plain"
  description = "Automated test resource for 'lambda-secrets'"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "plain" {
  secret_id     = aws_secretsmanager_secret.plain.id
  secret_string = var.secret-plain
}


resource "aws_secretsmanager_secret" "multiline" {
  name = "${local.secretsmanager-path}/multiline"
  description = "Automated test resource for 'lambda-secrets'"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "multiline" {
  secret_id     = aws_secretsmanager_secret.multiline.id
  secret_string = var.secret-multiline
}


resource "aws_secretsmanager_secret" "json" {
  name = "${local.secretsmanager-path}/json"
  description = "Automated test resource for 'lambda-secrets'"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "json" {
  secret_id     = aws_secretsmanager_secret.json.id
  secret_string = var.secret-json
}


resource "aws_secretsmanager_secret" "binary" {
  name = "${local.secretsmanager-path}/binary"
  description = "Automated test resource for 'lambda-secrets'"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "binary" {
  secret_id     = aws_secretsmanager_secret.binary.id
  secret_binary = var.secret-binary
}
