# AWS Lambda Secrets

Lambda extension layer to prefetch, persist and load secrets from AWS SecretsManager into Environment Variables for AWS Lambda Functions and Containers

## Features

* Prefetch secrets from AWS SecretsManager **only** on cold starts
* Persist secrets in the node's filesystem
* Load secrets into Env Vars before handler invocation

## Description

This repo implements two Lambda extensions delivered as 2 Go binaries (`fetch-secrets` and `load-secrets`) in a single Lambda Layer containing.

* `fetch-secrets` is an *external extension* which executes as a separate process on AWS Lambda execution environment cold starts ([Init Phase](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtime-environment.html#runtimes-lifecycle-ib)), reads SecretsManager ARNs or Paths and Environment Variable names from a YAML file, fetches the secret values and stores them in a `.env` file to be loaded as Environment Variables.

* `load-secrets` is an *internal extension* or *wrapper script* acting as an entrypoint before every Lambda invocation ([Invoke Phase](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtime-environment.html#runtimes-lifecycle-invoke)), which loads the secrets stored in the `.env` into Environment Variables making the available to the Lambda Application runtime before passing execution the the actual Handler.

[\1][\3]

## YAML File syntax

The YAML file can be located anywhere and the location can be set to the binary through the `-f` flag. Default is `secrets.yaml` in the working directory.

The syntax of the file is as follows:

```yaml
secrets: # a YAML list that contains maps of `valueFrom` and `name` keys
  - valueFrom: 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:aws-lambda-secrets/test1-VeF2Fu' # <-- Full ARN - Works!
    name: SECRET_VALUE_1
  - valueFrom: 'arn:aws:secretsmanager:eu-central-1:11111111111:secret:aws-lambda-secrets/test2' # <-- ARN without suffix - Works!
    name: SECRET_VALUE_2
  - valueFrom: 'aws-lambda-secrets/test3' # <-- Path for same AWS Account secrets - Works!
    name: SECRET_VALUE_3
  - valueFrom: '5up3r5ecre7p@55w0rd' # <-- The Secret itself - DOES NOT WORK!
    name: SECRET_VALUE_4
  [...]
```

The syntax is the YAML equivalent of the JSON [ECS Task Definition `secrets` field](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#secrets).

### Lambda Function Use-Case

To add the extension to a Lambda Function
1. Create a `secrets.yaml` file, following the above syntax in the Lambda application repo
2. Include the following Lambda Layer ARN to your Lambda function based on its architecture:
   * x86_64: `arn:aws:lambda:eu-central-1:533973265978:layer:aws-lambda-secrets-layer-x86_64:15`
   * arm64: `arn:aws:lambda:eu-central-1:533973265978:layer:aws-lambda-secrets-layer-arm64:15`
3. Set the value of the `AWS_LAMBDA_EXEC_WRAPPER` environment variable to `/opt/extensions/wrapper/load-secrets`

Terraform deployment [example](https://github.com/skroutz/aws-secretsmanager-lambda-example/blob/main/terraform/lambda-function.tf#L1):

```hcl
module "lambda-function-example" {
  source="terraform-aws-modules/lambda/aws"

  function_name   = "Function Name"
  description     = "Example Lambda Function with extension"
[...]
  layers = [
    "arn:aws:lambda:eu-central-1:533973265978:layer:aws-lambda-secrets-layer-x86_64:15"
  ]
[...]
  environment_variables = tomap({
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/extensions/wrapper/load-secrets"
  })
[...]
}
```
[\4][\5]

### Lambda Container Use-Case 

Lambda Container Dockerfile [example](https://github.com/skroutz/aws-secretsmanager-lambda-example/blob/main/lambda-container/Dockerfile):

```dockerfile
FROM public.ecr.aws/lambda/python:3.9

# Set LAMBDA_TASK_ROOT as the Lambda application directory
WORKDIR ${LAMBDA_TASK_ROOT}

# Add application files into LAMBDA_TASK_ROOT directory
COPY app.py ${LAMBDA_TASK_ROOT}/app.py
COPY secrets.yaml ${LAMBDA_TASK_ROOT}/secrets.yaml

# Add extension from aws-lambda-secrets-extension image into /opt
COPY --from=ghcr.io/skroutz/aws-lambda-secrets-extension:v1.0.0 /extension/fetch-secrets /opt/extensions/fetch-secrets
COPY --from=ghcr.io/skroutz/aws-lambda-secrets-extension:v1.0.0 /extension/wrapper/load-secrets /opt/extensions/wrapper/load-secrets

# Pass the actual ENTRYPOINT to '/opt/extensions/wrapper/load-secrets':
ENV AWS_LAMBDA_EXEC_WRAPPER "/opt/extensions/wrapper/load-secrets"

# Lambda Entrypoint CMD params
CMD ["app.lambda_handler"]
```

### Custom Lambda Container

Custom runtimes do not respect `AWS_LAMBDA_EXEC_WRAPPER`, thus will not execute the wrapper script as the function entrypoint. To enable wrapper scripts alongside [Runtime Interface Clients](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-images.html#runtimes-api-client), the extension script has to be set as the container's ENTRYPOINT responsible for passing execution to RIC executable with proper arguments, so that it invokes the function handler.

Lambda Custom Container for an [example](https://github.com/skroutz/aws-secretsmanager-lambda-example/tree/main/lambda-custom-container) Ruby Application

```dockerfile
FROM ruby:2.5

COPY secrets.yaml /var/task/secrets.yaml

COPY Gemfile /
COPY Gemfile.lock /

RUN apt-get -y update \
    && apt-get install -qqy \
    build-essential \
    && gem install bundler -v 1.17.3 \
    && bundle install \
    && rm -rf /var/lib/apt/lists/*

ENV AWS_LAMBDA_RUNTIME_API=3.9

COPY --from=ghcr.io/skroutz/aws-lambda-secrets-extension:v1.0.0 /extension/fetch-secrets /opt/extensions/fetch-secrets
COPY --from=ghcr.io/skroutz/aws-lambda-secrets-extension:v1.0.0 /extension/wrapper/load-secrets /opt/extensions/wrapper/load-secrets

COPY . /

ENTRYPOINT [ "/opt/extensions/wrapper/load-secrets" ]
CMD [ "/usr/local/bundle/bin/aws_lambda_ric", "app.Lambdas::App.process"]
```

For Github Actions CI integration, read access on the extension [package](https://github.com/skroutz/aws-lambda-secrets/pkgs/container/aws-lambda-secrets-extension) has to be granted for the Lambda App repository by the Security Team.

[\2][\5][\6]

## Reference
1. https://aws.amazon.com/blogs/compute/building-extensions-for-aws-lambda-in-preview/
2. https://aws.amazon.com/blogs/compute/working-with-lambda-layers-and-extensions-in-container-images/
3. https://developer.squareup.com/blog/using-aws-lambda-extensions-to-accelerate-aws-secrets-manager-access/
4. https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html#using-extensions-config
5. https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper
6. https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html#invocation-extensions-images

Previous Work and Relevant Repos
1. https://github.com/aws-samples/aws-lambda-extensions
2. https://github.com/hashicorp/vault-lambda-extension
3. https://github.com/square/lambda-secrets-prefetch
4. https://github.com/aws/aws-lambda-go