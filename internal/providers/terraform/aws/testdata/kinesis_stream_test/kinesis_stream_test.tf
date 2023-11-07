provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "aws" {
  alias                       = "ue2"
  region                      = "us-east-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for KinesisStream below

resource "aws_kinesis_stream" "test_stream_on_demand" {
  name = "terraform-kinesis-test-od"
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
  tags = {
    Environment = "test"
  }
}


resource "aws_kinesis_stream" "test_stream_on_demand_with_usage" {
  name = "terraform-kinesis-test-od-with-usage"
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
  tags = {
    Environment = "test"
  }
}

resource "aws_kinesis_stream" "test_stream_provisioned" {
  name = "terraform-kinesis-test-pr"
  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
  shard_count = 1
  tags = {
    Environment = "test"
  }
}

resource "aws_kinesis_stream" "test_stream_provisioned_with_usage" {
  name = "terraform-kinesis-test-with-usage"
  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
  shard_count = 4
  tags = {
    Environment = "test"
  }
}

resource "aws_kinesis_stream" "use2_test_stream_on_demand" {
  provider = aws.ue2
  name     = "use2_terraform-kinesis-test-od"
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
  tags = {
    Environment = "test"
  }
}


resource "aws_kinesis_stream" "use2_test_stream_on_demand_with_usage" {
  provider = aws.ue2
  name     = "use2_terraform-kinesis-test-od-with-usage"
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
  tags = {
    Environment = "test"
  }
}

resource "aws_kinesis_stream" "use2_test_stream_provisioned" {
  provider = aws.ue2
  name     = "use2_terraform-kinesis-test-pr"
  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
  shard_count = 1
  tags = {
    Environment = "test"
  }
}

resource "aws_kinesis_stream" "use2_test_stream_provisioned_with_usage" {
  provider = aws.ue2
  name     = "use2_terraform-kinesis-test-with-usage"
  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
  shard_count = 4
  tags = {
    Environment = "test"
  }
}

