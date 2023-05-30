provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_fsx_windows_file_system" "my_system" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "MULTI_AZ_1"
  storage_type        = "HDD"

  self_managed_active_directory {
    dns_ips     = ["10.0.0.111", "10.0.0.222"]
    domain_name = "corp.example.com"
    password    = "avoid-plaintext-passwords"
    username    = "Admin"
  }
}

resource "aws_fsx_windows_file_system" "my_file_system" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "MULTI_AZ_1"
  storage_type        = "HDD"

  self_managed_active_directory {
    dns_ips     = ["10.0.0.111", "10.0.0.222"]
    domain_name = "corp.example.com"
    password    = "avoid-plaintext-passwords"
    username    = "Admin"
  }
}

resource "aws_fsx_windows_file_system" "my_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "MULTI_AZ_1"
  storage_type        = "SSD"

  self_managed_active_directory {
    dns_ips     = ["10.0.0.111", "10.0.0.222"]
    domain_name = "corp.example.com"
    password    = "avoid-plaintext-passwords"
    username    = "Admin"
  }
}

resource "aws_fsx_windows_file_system" "my_file_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "MULTI_AZ_1"
  storage_type        = "SSD"

  self_managed_active_directory {
    dns_ips     = ["10.0.0.111", "10.0.0.222"]
    domain_name = "corp.example.com"
    password    = "avoid-plaintext-passwords"
    username    = "Admin"
  }
}
