module "lambda-container" {
  source = "github.com/terraform-aws-modules/terraform-aws-lambda?ref=v3.3.1"

  function_name = "${local.lambda-container-name}"
  description   = "Automated Test for ${local.project-name}"

  create_package = false
  create_role    = false

  package_type = "Image"
  image_uri = module.docker_image.image_uri

  create_lambda_function_url = true

  lambda_role = module.lambda_role.iam_role_arn

  depends_on = [module.docker_image]
}
