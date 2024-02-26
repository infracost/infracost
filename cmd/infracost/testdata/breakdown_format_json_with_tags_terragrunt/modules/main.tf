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
