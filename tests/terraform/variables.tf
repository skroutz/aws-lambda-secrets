variable "secret-plain" {
  default = "S3CR37 L1N3"
}

variable "secret-json" {
  default = <<EOF
{
  "username": "admin",
  "password": "p@55w0rd!"
}
EOF
}

variable "secret-multiline" {
  default = <<EOF
This is
a
Multiline
Secr3t!
EOF
}

variable "secret-binary" {
  default = "ewogICJ1c2VybmFtZSI6ICJhZG1pbiIsCiAgInBhc3N3b3JkIjogInBANTV3MHJkISIKICAiYmluYXJ5IjogdHJ1ZQp9Cg=="
}
