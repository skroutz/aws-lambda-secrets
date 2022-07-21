module "remote_state" {
  source  = "nozaq/remote-state-s3-backend/aws"
  version = "~>v1.2.0"

  providers = {
    aws = aws
    # Replication is disabled, we mock the replica provider alias
    aws.replica = aws
  }

  terraform_iam_policy_name_prefix = local.iam-resource-prefix

  dynamodb_table_name         = local.dynamodb-name
  dynamodb_table_billing_mode = "PAY_PER_REQUEST"

  s3_bucket_name          = local.s3-bucket-name
  override_s3_bucket_name = true

  kms_key_deletion_window_in_days = 7

  enable_replication = false

}
