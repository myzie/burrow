locals {
  name       = "burrow"
  account_id = data.aws_caller_identity.current.account_id
  lambda_policies = [
    aws_iam_policy.s3_user_policy.arn,
  ]
  timeout     = 60
  memory_size = 256
}

data "aws_caller_identity" "current" {}

// Virginia
module "region-us-east-1" {
  providers       = { aws = aws.us-east-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// California
module "region-us-west-1" {
  providers       = { aws = aws.us-west-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Oregon
module "region-us-west-2" {
  providers       = { aws = aws.us-west-2 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Dublin
module "region-eu-west-1" {
  providers       = { aws = aws.eu-west-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// London
module "region-eu-west-2" {
  providers       = { aws = aws.eu-west-2 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Ohio
module "region-us-east-2" {
  providers       = { aws = aws.us-east-2 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Sao Paulo
module "region-sa-east-1" {
  providers       = { aws = aws.sa-east-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Frankfurt
module "region-eu-central-1" {
  providers       = { aws = aws.eu-central-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Paris
module "region-eu-west-3" {
  providers       = { aws = aws.eu-west-3 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Stockholm
module "region-eu-north-1" {
  providers       = { aws = aws.eu-north-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Canada
module "region-ca-central-1" {
  providers       = { aws = aws.ca-central-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Seoul
module "region-ap-northeast-2" {
  providers       = { aws = aws.ap-northeast-2 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

// Mumbai
module "region-ap-south-1" {
  providers       = { aws = aws.ap-south-1 }
  source          = "../region"
  name            = local.name
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
  iam_policies    = local.lambda_policies
  timeout         = local.timeout
  memory_size     = local.memory_size
}

# S3 policy
resource "aws_iam_policy" "s3_user_policy" {
  name        = "burrow-s3-${local.name}"
  description = "burrow-s3-${local.name}"
  path        = "/"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketLocation",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:DeleteObjectVersion",
        ]
        Effect = "Allow"
        Resource = [
          "arn:aws:s3:::${local.name}-${local.account_id}-us-east-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-east-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-west-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-west-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-2",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-2/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-west-2",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-west-2/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-east-2",
          "arn:aws:s3:::${local.name}-${local.account_id}-us-east-2/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-sa-east-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-sa-east-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-central-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-central-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-3",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-west-3/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-north-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-eu-north-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-ca-central-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-ca-central-1/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-ap-northeast-2",
          "arn:aws:s3:::${local.name}-${local.account_id}-ap-northeast-2/*",
          "arn:aws:s3:::${local.name}-${local.account_id}-ap-south-1",
          "arn:aws:s3:::${local.name}-${local.account_id}-ap-south-1/*",
        ]
      },
    ]
  })
}
