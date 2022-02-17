resource "random_password" "front_db_password" {
  length  = 64
  special = false
}

resource "aws_db_instance" "front" {
  identifier                = "${var.env_name}-front"
  allocated_storage         = 100
  engine                    = "mysql"
  instance_class            = "db.t3.small"
  db_subnet_group_name      = var.database_subnet_group_name
  vpc_security_group_ids    = var.database_security_group_ids
  name                      = "front"
  username                  = "front"
  password                  = random_password.front_db_password.result
  skip_final_snapshot       = var.env_name == "prod"
  final_snapshot_identifier = "front-db-final"

  tags = {
    Environment = var.env_name
  }
}

output "front_db_url" {
  value = "mysql://front:${random_password.front_db_password.result}@${aws_db_instance.front.endpoint}/front"
}

output "front_db_password" {
  value = random_password.front_db_password.result
}
