data "aws_region" "current" {}

output "full_name" {
  value = data.aws_region.current.name
}
