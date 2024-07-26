
data "aws_region" "current" {}

resource "aws_lambda_function" "this" {
  function_name    = var.name
  description      = var.description
  role             = aws_iam_role.this.arn
  handler          = var.handler
  memory_size      = var.memory_size
  runtime          = var.runtime
  timeout          = var.timeout
  kms_key_arn      = var.kms_key
  architectures    = ["x86_64"]
  filename         = var.zip_path
  source_code_hash = filebase64sha256(var.zip_path)
  tags             = var.tags
  tracing_config {
    mode = "Active"
  }
  environment {
    variables = var.environment
  }
  depends_on = [
    aws_cloudwatch_log_group.lambda,
    aws_iam_role_policy_attachment.vpc_access,
    aws_iam_role_policy_attachment.eni_mgmt,
  ]
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.name}"
  retention_in_days = 365
  kms_key_id        = var.kms_key
  tags              = var.tags
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "this" {
  name               = "${var.name}-${data.aws_region.current.name}"
  description        = "${var.name}-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "vpc_access" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy_attachment" "eni_mgmt" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaENIManagementAccess"
}

resource "aws_iam_role_policy_attachment" "extra_policies" {
  count      = length(var.iam_policies)
  role       = aws_iam_role.this.name
  policy_arn = var.iam_policies[count.index]
}
