variable "bucket_name" {
  description = "Name of the S3 bucket"
  type        = string
}

variable "log_bucket" {
  description = "Name of the bucket to store access logs"
  type        = string
  default     = null
}

variable "tags" {
  description = "Tags to be applied to the bucket"
  type        = map(string)
  default     = {}
}
