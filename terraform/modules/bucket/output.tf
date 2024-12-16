
output "bucket_arn" {
  value = aws_s3_bucket.this.arn
}

output "bucket_name" {
  value = aws_s3_bucket.this.bucket
}

output "bucket_region" {
  value = aws_s3_bucket.this.region
}

output "read_policy_arn" {
  value = aws_iam_policy.read_policy.arn
}

output "write_policy_arn" {
  value = aws_iam_policy.write_policy.arn
}
