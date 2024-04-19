provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "valid" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800
  }
}

resource "aws_instance" "ebs_invalid" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "invalid"
    volume_size = 1000
    iops        = 800
  }
}

resource "aws_instance" "instance_invalid" {
  ami           = "ami-674cbc1e"
  instance_type = "invalid_instance_type"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800
  }
}

resource "aws_db_instance" "valid" {
  engine         = "mysql"
  instance_class = "db.t3.large"
}

resource "aws_db_instance" "invalid" {
  engine         = "mysql"
  instance_class = "invalid_instance_class"
}

