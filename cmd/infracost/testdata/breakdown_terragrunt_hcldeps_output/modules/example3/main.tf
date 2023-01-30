data "aws_secretsmanager_secret" "bad" {
  name = "example"
}

output "retrieved_secrets" {
  value     = values(data.aws_secretsmanager_secret.bad)[*].value
  sensitive = true
}
