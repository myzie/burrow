variable "name" {
  description = "Function name"
  type        = string
}

variable "handler" {
  description = "Function entrypoint"
  type        = string
}

variable "runtime" {
  description = "Function runtime"
  type        = string
  default     = "provided.al2023"
}

variable "kms_key_arn" {
  description = "KMS key to use for encryption operations"
  type        = string
  default     = null
}

variable "memory_size" {
  description = "Memory size in MB to assign to the function"
  type        = number
  default     = 256
}

variable "timeout" {
  description = "Function invocation timeout in seconds"
  type        = number
  default     = 10
}

variable "environment" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "tags" {
  description = "Tags to assign to the function"
  type        = map(string)
  default     = {}
}

variable "filename" {
  description = "Path to the function code on the local filesystem"
  type        = string
}

variable "architectures" {
  description = "List of architectures to build the function for"
  type        = list(string)
  default     = ["x86_64"]
}

variable "iam_role_arn" {
  description = "IAM role ARN to assign to the function"
  type        = string
}

variable "log_retention" {
  description = "Log group retention in days"
  type        = number
  default     = 365
}

variable "authorization_type" {
  description = "Authorization type for the function URL"
  type        = string
  default     = "NONE"
}

variable "cors" {
  description = "CORS configuration for the function URL"
  type = object({
    allow_credentials = bool
    allow_origins     = list(string)
    allow_methods     = list(string)
    allow_headers     = list(string)
    expose_headers    = list(string)
    max_age           = number
  })
  default = null
}

variable "bucket_name" {
  description = "Name of the S3 bucket to use for storage"
  type        = string
  default     = null
}

variable "bucket_region" {
  description = "Region where the S3 bucket resides"
  type        = string
  default     = null
}
