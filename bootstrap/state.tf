terraform {
  backend "s3" {

    encrypt        = true

    # Values as generated from 'bootstrap/'
    region         = "eu-central-1"
    bucket         = "lambda-secrets-app-state-533973265978"
    dynamodb_table = "lambda-secrets-app-state"
    kms_key_id     = "b846806b-b83f-4ac5-b3b0-92505b31bdeb"

    key            = "tf-state/aws-lambda-secrets/bootstrap.tfstate"
  }
}