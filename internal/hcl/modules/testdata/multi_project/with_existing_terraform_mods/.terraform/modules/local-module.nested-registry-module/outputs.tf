output "sns_topic_arn" {
  description = "ARN of SNS topic"
  value       = element(concat(aws_sns_topic.this.*.arn, [""]), 0)
}
