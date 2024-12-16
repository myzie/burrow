
terraform {
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      configuration_aliases = [aws]
    }
  }
}

# Bucket
resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name
  tags   = var.tags
}

# Ownership controls
resource "aws_s3_bucket_ownership_controls" "this" {
  bucket = aws_s3_bucket.this.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# Bucket ACL
resource "aws_s3_bucket_acl" "this" {
  depends_on = [aws_s3_bucket_ownership_controls.this]
  bucket     = aws_s3_bucket.this.id
  acl        = "private"
}

# SSE encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Logging configuration
resource "aws_s3_bucket_logging" "this" {
  count         = var.log_bucket != null ? 1 : 0
  bucket        = aws_s3_bucket.this.id
  target_bucket = var.log_bucket
  target_prefix = "log/${var.bucket_name}/"
}

# Public access block
resource "aws_s3_bucket_public_access_block" "this" {
  bucket                  = aws_s3_bucket.this.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Bucket policy
resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "require-https"
    Statement = [
      {
        Sid       = "DenyNonSSLRequests"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.this.arn,
          "${aws_s3_bucket.this.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      }
    ]
  })
}

# Read-only access policy
resource "aws_iam_policy" "read_policy" {
  name_prefix = "s3-bucket-read-"
  path        = "/"
  description = "Read permissions for ${aws_s3_bucket.this.id}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject*",
        ]
        Effect   = "Allow"
        Resource = "${aws_s3_bucket.this.arn}/*",
      },
      {
        Action = [
          "s3:ListBucket",
        ]
        Effect   = "Allow"
        Resource = "${aws_s3_bucket.this.arn}",
      },
    ]
  })
}

# Read/write access policy
resource "aws_iam_policy" "write_policy" {
  name_prefix = "s3-bucket-write-"
  path        = "/"
  description = "Read/write permissions for ${aws_s3_bucket.this.id}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject*",
          "s3:PutObject",
          "s3:PutObjectAcl",
          "s3:DeleteObject",
        ]
        Effect   = "Allow"
        Resource = "${aws_s3_bucket.this.arn}/*",
      },
      {
        Action = [
          "s3:ListBucket",
        ]
        Effect   = "Allow"
        Resource = "${aws_s3_bucket.this.arn}",
      },
    ]
  })
}
