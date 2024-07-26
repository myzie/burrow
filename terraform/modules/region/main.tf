
data "aws_caller_identity" "current" {}

module "burrow" {
  source       = "../lambda"
  name         = var.name
  revision     = var.revision
  description  = "Lambda for ${var.name}"
  runtime      = "provided.al2023"
  handler      = "${var.name}-eval"
  zip_path     = var.lambda_zip_path
  timeout      = var.timeout
  memory_size  = var.memory_size
  iam_policies = var.iam_policies
}
