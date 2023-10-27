provider "aws" {
  region                      = "eu-west-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}
resource "aws_kinesis_firehose_delivery_stream" "withAllTags" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"
  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = "fake"
        role_arn      = "fake"
        table_name    = "fake"
      }
    }
  }

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    vpc_config {
      security_group_ids = ["fake"]
      subnet_ids         = ["fake", "fake1"]
      role_arn           = aws_iam_role.firehose.arn
    }

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket"
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_iam_role" "firehose" {
  name = "firehose_test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test"
}

resource "aws_kinesis_firehose_delivery_stream" "EnabledFalse" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    vpc_config {
      security_group_ids = ["fake"]
      subnet_ids         = ["fake", "fake1"]
      role_arn           = aws_iam_role.firehose.arn
    }

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "onlyDataIngested" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "withoutUsage" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"
  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = "fake"
        role_arn      = "fake"
        table_name    = "fake"
      }
    }
  }

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    vpc_config {
      security_group_ids = ["fake"]
      subnet_ids         = ["fake", "fake1"]
      role_arn           = aws_iam_role.firehose.arn
    }

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "forTwoMilGB" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test2" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.2.0/24"
}

resource "aws_kinesis_firehose_delivery_stream" "with_dynamic_subnet" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"
  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = "fake"
        role_arn      = "fake"
        table_name    = "fake"
      }
    }
  }

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    vpc_config {
      security_group_ids = ["fake"]
      subnet_ids = [
        aws_subnet.test.id,
        aws_subnet.test2.id
      ]
      role_arn = aws_iam_role.firehose.arn
    }

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }

  splunk_configuration {
    hec_endpoint = "fake"
    hec_token    = "fake"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
