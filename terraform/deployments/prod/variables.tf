
variable "lambda_zip_path" {
  description = "Path to the function code on the local filesystem"
  type        = string
}

variable "revision" {
  description = "Function code revision"
  type        = string
}

variable "bucket_name" {
  description = "Name of the S3 bucket to store the function code"
  type        = string
}

variable "bucket_key" {
  description = "Key of the S3 object to store the function code"
  type        = string
}
