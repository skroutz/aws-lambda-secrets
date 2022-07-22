resource "aws_secretsmanager_secret" "plain" {
  name = "${local.secretsmanager-path}/simple"
}

resource "aws_secretsmanager_secret_version" "plain" {
  secret_id     = aws_secretsmanager_secret.plain.id
  secret_string = "S3CR37 L1N3"
}


resource "aws_secretsmanager_secret" "json" {
  name = "${local.secretsmanager-path}/json"
}

resource "aws_secretsmanager_secret_version" "json" {
  secret_id     = aws_secretsmanager_secret.json.id
  secret_string = <<EOF
{
  "username": "admin",
  "password": "p@55w0rd!"
}
EOF
}

resource "aws_secretsmanager_secret" "binary" {
  name = "${local.secretsmanager-path}/binary"
}

resource "aws_secretsmanager_secret_version" "binary" {
  secret_id     = aws_secretsmanager_secret.binary.id
  secret_binary = base64encode(<<EOF
{
  "username": "admin",
  "password": "p@55w0rd!"
}
EOF
)
}
