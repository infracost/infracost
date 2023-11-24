provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  dynamic "ebs_block_device" {
    for_each = {
      device1 = {
        name = "device1"
        options = {
          volume_size = 1000
          iops        = 800
        }
      }
      device2 = {
        name = "device2"
        options = {
          volume_size = 500
          iops        = 400
        }
      }
    }
    iterator = device
    content {
      device_name = device.value.name
      volume_type = "io1"
      volume_size = device.value.options.volume_size
      iops        = device.value.options.iops
    }
  }
}

resource "aws_instance" "web_app2" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  dynamic "ebs_block_device" {
    for_each = {
      device1 = {
        name = "device1"
        options = {
          volume_size = 1000
          iops        = 800
        }
      }
      device2 = {
        name = "device2"
        options = {
          volume_size = 500
          iops        = 400
        }
      }
    }
    content {
      device_name = ebs_block_device.value.name
      volume_type = "io1"
      volume_size = ebs_block_device.value.options.volume_size
      iops        = ebs_block_device.value.options.iops
    }
  }

  dynamic "ebs_block_device" {
    for_each = {}
    content {
      device_name = ebs_block_device.value.name
      volume_type = "io1"
      volume_size = ebs_block_device.value.options.volume_size
      iops        = ebs_block_device.value.options.iops
    }
  }
}

