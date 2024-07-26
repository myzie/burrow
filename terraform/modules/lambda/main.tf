
resource "aws_lambda_function" "lambda" {
  architectures    = var.architectures
  filename         = var.filename
  function_name    = var.name
  description      = "${var.name} function"
  handler          = var.handler
  kms_key_arn      = var.kms_key_arn
  memory_size      = var.memory_size
  role             = var.iam_role_arn
  runtime          = var.runtime
  source_code_hash = filebase64sha256(var.filename)
  tags             = var.tags
  timeout          = var.timeout
  tracing_config {
    mode = "Active"
  }
  environment {
    variables = var.environment
  }
  depends_on = [
    aws_cloudwatch_log_group.lambda
  ]
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.name}"
  retention_in_days = var.log_retention
  kms_key_id        = var.kms_key_arn
  tags              = var.tags
}

resource "aws_lambda_function_url" "lambda" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = var.authorization_type
  dynamic "cors" {
    for_each = var.cors != null ? [var.cors] : []
    content {
      allow_credentials = cors.value.allow_credentials
      allow_origins     = cors.value.allow_origins
      allow_methods     = cors.value.allow_methods
      allow_headers     = cors.value.allow_headers
      expose_headers    = cors.value.expose_headers
      max_age           = cors.value.max_age
    }
  }
}
