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

resource "aws_autoscaling_group" "bar" {
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true

  tag {
    key                 = "foo"
    value               = "bar"
    propagate_at_launch = true
  }

  tag {
    key                 = "lorem"
    value               = "ipsum"
    propagate_at_launch = false
  }
}
