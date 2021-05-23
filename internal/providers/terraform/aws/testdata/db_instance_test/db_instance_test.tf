provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_db_instance" "mysql-default" {
  engine         = "mysql"
  instance_class = "db.t3.large"
}

resource "aws_db_instance" "mysql-allocated-storage" {
  engine            = "mysql"
  instance_class    = "db.t3.large"
  allocated_storage = 20
}

resource "aws_db_instance" "mysql-multi-az" {
  engine            = "mysql"
  instance_class    = "db.t3.large"
  multi_az          = true
  allocated_storage = 30
}

resource "aws_db_instance" "mysql-magnetic" {
  engine            = "mysql"
  instance_class    = "db.t3.large"
  storage_type      = "standard"
  allocated_storage = 40
}

resource "aws_db_instance" "mysql-iops" {
  engine            = "mysql"
  instance_class    = "db.t3.large"
  storage_type      = "io1"
  allocated_storage = 50
  iops              = 500
}

resource "aws_db_instance" "aurora" {
  engine         = "aurora"
  instance_class = "db.t3.small"
}
resource "aws_db_instance" "aurora-mysql" {
  engine         = "aurora-mysql"
  instance_class = "db.t3.small"
}
resource "aws_db_instance" "aurora-postgresql" {
  engine         = "aurora-postgresql"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "mariadb" {
  engine         = "mariadb"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "mysql" {
  engine         = "mysql"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "postgres" {
  engine         = "postgres"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "oracle-se" {
  engine         = "oracle-se"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "oracle-se1" {
  engine         = "oracle-se1"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "oracle-se2" {
  engine         = "oracle-se2"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "oracle-ee" {
  engine         = "oracle-ee"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "sqlserver-ex" {
  engine         = "sqlserver-ex"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "sqlserver-web" {
  engine         = "sqlserver-web"
  instance_class = "db.t3.large"
}
resource "aws_db_instance" "sqlserver-se" {
  engine         = "sqlserver-se"
  instance_class = "db.m5.xlarge"
}
resource "aws_db_instance" "sqlserver-ee" {
  engine         = "sqlserver-ee"
  instance_class = "db.m5.xlarge"
}

resource "aws_db_instance" "oracle-se1-byol" {
  engine         = "oracle-se1"
  instance_class = "db.t3.large"
  license_model  = "bring-your-own-license"
}
