
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
  name               = "${var.name}-role"
  description        = "${var.name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "xray_access" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}

resource "aws_iam_role_policy_attachment" "logs_access" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

// Virginia
module "region-us-east-1" {
  providers     = { aws = aws.us-east-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// California
module "region-us-west-1" {
  providers     = { aws = aws.us-west-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Oregon
module "region-us-west-2" {
  providers     = { aws = aws.us-west-2 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Dublin
module "region-eu-west-1" {
  providers     = { aws = aws.eu-west-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// London
module "region-eu-west-2" {
  providers     = { aws = aws.eu-west-2 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Ohio
module "region-us-east-2" {
  providers     = { aws = aws.us-east-2 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Sao Paulo
module "region-sa-east-1" {
  providers     = { aws = aws.sa-east-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Frankfurt
module "region-eu-central-1" {
  providers     = { aws = aws.eu-central-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Paris
module "region-eu-west-3" {
  providers     = { aws = aws.eu-west-3 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Stockholm
module "region-eu-north-1" {
  providers     = { aws = aws.eu-north-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Canada
module "region-ca-central-1" {
  providers     = { aws = aws.ca-central-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Seoul
module "region-ap-northeast-2" {
  providers     = { aws = aws.ap-northeast-2 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}

// Mumbai
module "region-ap-south-1" {
  providers     = { aws = aws.ap-south-1 }
  source        = "../modules/lambda"
  log_retention = var.log_retention
  name          = var.name
  tags          = var.tags
  filename      = var.lambda_filename
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  architectures = var.lambda_architectures
  iam_role_arn  = aws_iam_role.this.arn
}
