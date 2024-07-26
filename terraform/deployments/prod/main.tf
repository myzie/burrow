
module "burrow" {
  providers       = { aws = aws.us-east-1 }
  source          = "../../modules/main"
  lambda_zip_path = var.lambda_zip_path
  bucket_name     = var.bucket_name
  bucket_key      = var.bucket_key
  revision        = var.revision
}
