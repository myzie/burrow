
variable "name" {
  description = "Application name"
  type        = string
}

variable "regions" {
  description = "Deployment regions"
  type        = list(string)
  default = [
    "us-east-1",
    "us-east-2",
    "us-west-1",
    "us-west-2",
    "ca-central-1",
  ]
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "log_retention" {
  description = "Log group retention in days"
  type        = number
  default     = 365
}

variable "git_revision" {
  description = "Git revision"
  type        = string
  default     = ""
}

variable "lambda_filename" {
  description = "Lambda filename"
  type        = string
  default     = ""
}

variable "lambda_handler" {
  description = "Lambda handler"
  type        = string
  default     = ""
}

variable "lambda_architectures" {
  description = "Lambda architectures"
  type        = list(string)
  default     = ["x86_64"]
}

variable "lambda_runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "provided.al2023"
}

variable "lambda_memory_size" {
  description = "Lambda memory size"
  type        = number
  default     = 256
}

variable "lambda_timeout" {
  description = "Lambda timeout"
  type        = number
  default     = 10
}

variable "bucket_name" {
  description = "Name of the S3 bucket to create"
  type        = string
}

variable "log_bucket" {
  description = "Name of the bucket to store access logs"
  type        = string
  default     = null
}
