resource "aws_dynamodb_table" "sessions" {
  name           = "${var.env_name}-sessions"
  billing_mode   = "PROVISIONED"
  read_capacity  = 50
  write_capacity = 50
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  ttl {
    attribute_name = "expires"
    enabled        = true
  }

  point_in_time_recovery {
    enabled = true
  }
}

resource "random_password" "app_secret" {
  length           = 64
  special          = true
  override_special = "!#*?"
}
