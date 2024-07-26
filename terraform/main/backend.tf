
# This file is used to configure the backend for the terraform state file.
# The S3 bucket (and optionally the key and region) should be overriden with
# your desired values when running "terraform init". For example:
#
# terraform init -backend-config=my-bucket-1234

terraform {
  backend "s3" {
    bucket = ""
    key    = "states/burrow/terraform.tfstate"
    region = "us-east-1"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60.0"
    }
  }
}
