# backend.tf
terraform {
  backend "s3" {
    endpoint = "https://fsn1.your-objectstorage.com"
    bucket   = "minecraft-tfstate"
    key      = "spinoff/terraform/environments/prod/terraform.tfstate"
    region   = "fsn1"

    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    force_path_style            = true
  }
}
