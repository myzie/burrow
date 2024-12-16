output "function_urls" {
  value = {
    "us-east-1"      = try(module.region-us-east-1[0].url, null)
    "us-east-2"      = try(module.region-us-east-2[0].url, null)
    "us-west-1"      = try(module.region-us-west-1[0].url, null)
    "us-west-2"      = try(module.region-us-west-2[0].url, null)
    "eu-west-1"      = try(module.region-eu-west-1[0].url, null)
    "eu-west-2"      = try(module.region-eu-west-2[0].url, null)
    "eu-west-3"      = try(module.region-eu-west-3[0].url, null)
    "eu-central-1"   = try(module.region-eu-central-1[0].url, null)
    "sa-east-1"      = try(module.region-sa-east-1[0].url, null)
    "eu-north-1"     = try(module.region-eu-north-1[0].url, null)
    "ca-central-1"   = try(module.region-ca-central-1[0].url, null)
    "ap-south-1"     = try(module.region-ap-south-1[0].url, null)
    "ap-northeast-1" = try(module.region-ap-northeast-1[0].url, null)
    "ap-northeast-2" = try(module.region-ap-northeast-2[0].url, null)
    "ap-northeast-3" = try(module.region-ap-northeast-3[0].url, null)
    "ap-southeast-1" = try(module.region-ap-southeast-1[0].url, null)
    "ap-southeast-2" = try(module.region-ap-southeast-2[0].url, null)
  }
}
