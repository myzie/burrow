
variable "name" {
  description = "Application name"
  type        = string
}

variable "regions" {
  description = "Deployment regions"
  type        = list(string)
  default = [
    "us-east-1",
    "us-west-1",
    "us-west-2",
    "eu-west-1",
    "eu-west-2",
    "us-east-2",
    "sa-east-1",
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
