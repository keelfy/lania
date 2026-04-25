terraform {
  backend "s3" {
    endpoints = {
      s3 = "https://fsn1.your-objectstorage.com"
    }

    bucket = "lania"
    key    = "minecraft/terraform.tfstate"
    region = "fsn1"

    skip_credentials_validation  = true
    skip_metadata_api_check      = true
    skip_region_validation       = true
    use_path_style             = true
    skip_requesting_account_id   = true
  }
}