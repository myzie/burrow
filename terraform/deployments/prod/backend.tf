
terraform {
  backend "s3" {
    bucket = "override-me"
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
