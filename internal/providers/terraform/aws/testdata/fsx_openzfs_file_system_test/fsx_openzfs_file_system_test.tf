provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_fsx_openzfs_file_system" "my_simple_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "SINGLE_AZ_1"
  storage_type        = "SSD"
}

resource "aws_fsx_openzfs_file_system" "my_iops_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "SINGLE_AZ_1"
  storage_type        = "SSD"
  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = 1000
  }
}

resource "aws_fsx_openzfs_file_system" "my_compressed_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "SINGLE_AZ_1"
  storage_type        = "SSD"
  root_volume_configuration {
    data_compression_type = "LZ4"
  }
}

resource "aws_fsx_openzfs_file_system" "my_compressed_default_system_ssd" {
  storage_capacity    = 300
  subnet_ids          = ["fake"]
  throughput_capacity = 1024
  deployment_type     = "SINGLE_AZ_1"
  storage_type        = "SSD"
  root_volume_configuration {
    data_compression_type = "LZ4"
  }
}
