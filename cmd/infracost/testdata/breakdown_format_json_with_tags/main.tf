provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride"
      DefaultOverride    = "defaultoverride"
    }
  }
}

resource "aws_sns_topic_subscription" "sns_topic_noTags" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}

resource "aws_sns_topic_subscription" "sns_topic_withTags" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
  tags = {
    DefaultOverride = "sns-def"
    ResourceTag     = "sns-ghi"
  }
}

resource "aws_sqs_queue" "sqs_noTags" {
  name = "sqs_noTags"
}

resource "aws_sqs_queue" "sqs_withTags" {
  name = "sqs_withTags"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_disk" "gcd1" {
  name = "gcd1"
  type = "pd-standard"

  labels = {
    GoogleLabel = "compute-disk-label"
  }
}

resource "google_compute_disk" "gcd2" {
  name = "gcd2"
  type = "pd-ssd"
}

provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_redis_cache" "arc1" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 3
  family              = "C"
  sku_name            = "Standard"
  tags = {
    AzureLabel = "azure-tag"
  }
}

resource "azurerm_redis_cache" "arc2" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 3
  family              = "C"
  sku_name            = "Standard"
}