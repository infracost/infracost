provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_codebuild_project" "my_project_noUsage" {
  name        = "test-project-cache"
  description = "test_codebuild_project_cache"

  service_role = ""

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_MEDIUM"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  source {
    type            = "GITHUB"
    location        = ""
    git_clone_depth = 1
  }
}

resource "aws_codebuild_project" "my_small_project" {
  name        = "test-project-cache"
  description = "test_codebuild_project_cache"

  service_role = ""

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  source {
    type            = "GITHUB"
    location        = ""
    git_clone_depth = 1
  }
}

resource "aws_codebuild_project" "my_medium_project" {
  name        = "test-project-cache"
  description = "test_codebuild_project_cache"

  service_role = ""

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_MEDIUM"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  source {
    type            = "GITHUB"
    location        = ""
    git_clone_depth = 1
  }
}

resource "aws_codebuild_project" "my_large_linux_project" {
  name        = "test-project-cache"
  description = "test_codebuild_project_cache"

  service_role = ""

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_LARGE"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  source {
    type            = "GITHUB"
    location        = ""
    git_clone_depth = 1
  }
}

resource "aws_codebuild_project" "my_large_windows_project" {
  name        = "test-project-cache"
  description = "test_codebuild_project_cache"

  service_role = ""

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_LARGE"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "WINDOWS_SERVER_2019_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  source {
    type            = "GITHUB"
    location        = ""
    git_clone_depth = 1
  }
}
