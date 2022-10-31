
output "aws-deployer" {

  description = "The IAM API Keys to set in Github Secrets"
  sensitive   = true
  value = {
    "AWS_ACCESS_KEY_ID" : module.user.iam_access_key_id,
    "AWS_SECRET_ACCESS_KEY" : module.user.iam_access_key_secret,
    "ASSUME_ROLE" : local.iam-deployer-role,
  }
}


output "tf-state-resources" {

  description = "The values to be set with Github Workflow"
  value = {
    "AWS_REGION" : data.aws_region.current.name
    "BUCKET_NAME" : module.remote_state.state_bucket.bucket
    "DYNAMODB_NAME" : module.remote_state.dynamodb_table.id
    "KMS_KEY_NAME" : module.remote_state.kms_key.id
  }
}

