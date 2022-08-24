terraform {
  backend "s3" {

    encrypt        = true

    # Values as generated from 'bootstrap/'
    region         = "eu-central-1"
    bucket         = "aws-lambda-secrets-layer-state-533973265978"
    dynamodb_table = "aws-lambda-secrets-layer-state"
    kms_key_id     = "e740f26a-7474-4774-9302-583f98aed23f"

    key            = "tf-state/aws-lambda-secrets/terraform/terraform.tfstate"
  }
}
