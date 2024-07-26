
variable "name" {
  description = "Function name"
  type        = string
  default     = ""
}

variable "handler" {
  description = "Function entrypoint"
  type        = string
  default     = ""
}

variable "runtime" {
  description = "Function runtime"
  type        = string
  default     = ""
}

variable "description" {
  description = "Function description"
  type        = string
  default     = ""
}

variable "kms_key" {
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
  default     = 5
}

variable "environment" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "tracing_mode" {
  description = "Function tracing mode"
  type        = string
  default     = null
}

variable "tags" {
  description = "Tags to assign to the function"
  type        = map(string)
  default     = {}
}

variable "zip_path" {
  description = "Path to the function code on the local filesystem"
  type        = string
}

variable "revision" {
  description = "Function code revision"
  type        = string
}

variable "iam_policies" {
  description = "List of IAM policies to attach to the function role"
  type        = list(string)
  default     = []
}
