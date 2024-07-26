
output "function_urls" {
  value = {
    "us-east-1"      = module.region-us-east-1.url
    "us-east-2"      = module.region-us-east-2.url
    "us-west-1"      = module.region-us-west-1.url
    "us-west-2"      = module.region-us-west-2.url
    "eu-west-1"      = module.region-eu-west-1.url
    "eu-west-2"      = module.region-eu-west-2.url
    "eu-west-3"      = module.region-eu-west-3.url
    "eu-central-1"   = module.region-eu-central-1.url
    "sa-east-1"      = module.region-sa-east-1.url
    "eu-north-1"     = module.region-eu-north-1.url
    "ca-central-1"   = module.region-ca-central-1.url
    "ap-south-1"     = module.region-ap-south-1.url
    "ap-northeast-1" = module.region-ap-northeast-1.url
    "ap-northeast-2" = module.region-ap-northeast-2.url
    "ap-northeast-3" = module.region-ap-northeast-3.url
    "ap-southeast-1" = module.region-ap-southeast-1.url
    "ap-southeast-2" = module.region-ap-southeast-2.url
  }
}
