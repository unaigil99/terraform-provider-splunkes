# Configure the Splunk Enterprise Security provider
terraform {
  required_providers {
    splunkes = {
      source  = "Agentric/splunkes"
      version = "~> 1.0"
    }
  }
}

provider "splunkes" {
  url                  = "https://splunk.example.com:8089"
  username             = var.splunk_username
  password             = var.splunk_password
  insecure_skip_verify = true
  timeout              = 60
}

# Alternatively, use environment variables:
# SPLUNK_URL, SPLUNK_USERNAME, SPLUNK_PASSWORD, SPLUNK_AUTH_TOKEN

# Or use a bearer token:
# provider "splunkes" {
#   url        = "https://splunk.example.com:8089"
#   auth_token = var.splunk_auth_token
# }

variable "splunk_username" {
  type      = string
  sensitive = true
}

variable "splunk_password" {
  type      = string
  sensitive = true
}
