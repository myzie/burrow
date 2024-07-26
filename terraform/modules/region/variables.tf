
variable "name" {
  description = "Infrastructure name"
  type        = string
}

variable "region_index" {
  description = "Region index"
  type        = number
  default     = 1
}

variable "bucket_name" {
  description = "S3 bucket for the function code"
  type        = string
}

variable "bucket_key" {
  description = "S3 key for the function code"
  type        = string
}

variable "lambda_zip_path" {
  description = "Path to the function code on the local filesystem"
  type        = string
}

variable "iam_policies" {
  description = "List of IAM policies to attach to the function role"
  type        = list(string)
  default     = []
}

variable "revision" {
  description = "Function code revision"
  type        = string
}

variable "memory_size" {
  description = "Memory size in MB to assign to the function"
  type        = number
  default     = 256
}

variable "timeout" {
  description = "Function invocation timeout in seconds"
  type        = number
  default     = 5
}
