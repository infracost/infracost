provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800
  }
}

# this resource is not supported
resource "aws_codepipeline" "codepipeline" {
  name     = "tf-test-pipeline"
  role_arn = "fake"

  artifact_store {
    location = "bucket"
    type     = "S3"
  }

  stage {
    name = "Source-changed"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["source_output"]

      configuration = {
        ConnectionArn    = "fake"
        FullRepositoryId = "my-organization/example"
        BranchName       = "main"
      }
    }
  }
}

# this resource is marked as free
resource "aws_codebuild_webhook" "example" {
  project_name = "test"
  build_type   = "BUILD"
  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PUSH"
    }

    filter {
      type    = "BASE_REF"
      pattern = "master"
    }
  }
}
