provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNoOverride = "default_no_override"
      DefaultOverride   = "default_override"
    }
  }
}

resource "aws_instance" "with_implicit_root_block_device" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
}

resource "aws_instance" "with_implicit_root_block_device_and_volume_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  volume_tags = {
    VolTag          = "volume_tag"
    DefaultOverride = "volume_tag_override"
  }
}

resource "aws_instance" "with_implicit_root_block_device_and_ebs_block_device_no_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
  }
}

resource "aws_instance" "with_implicit_root_block_device_and_ebs_block_device_no_tags_and_volume_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
  }

  volume_tags = {
    VolTag          = "volume_tag"
    DefaultOverride = "volume_tag_override"
  }
}

resource "aws_instance" "with_implicit_root_block_device_and_ebs_block_device_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
    tags = {
      EBDTagKey       = "ebd_tag_val"
      DefaultOverride = "ebd_tag_override"
    }
  }
}

resource "aws_instance" "with_root_block_device_no_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }
}

resource "aws_instance" "with_root_block_device_no_tags_and_volume_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  volume_tags = {
    VolTag          = "volume_tag"
    DefaultOverride = "volume_tag_override"
  }
}

resource "aws_instance" "with_root_block_device_no_tags_and_ebs_block_device_no_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
  }
}

resource "aws_instance" "with_root_block_device_no_tags_and_ebs_block_device_no_tags_and_volume_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
  }

  volume_tags = {
    VolTag          = "volume_tag"
    DefaultOverride = "volume_tag_override"
  }
}

resource "aws_instance" "with_root_block_device_no_tags_and_ebs_block_device_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
    tags = {
      EBDTagKey       = "ebd_tag_val"
      DefaultOverride = "ebd_tag_override"
    }
  }
}

resource "aws_instance" "with_root_block_device_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50

    tags = {
      RBDTagKey       = "rbd_tag_val"
      DefaultOverride = "rbd_tag_override"
    }
  }
}

resource "aws_instance" "with_root_block_device_tags_and_ebs_block_device_no_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50

    tags = {
      RBDTagKey       = "rbd_tag_val"
      DefaultOverride = "rbd_tag_override"
    }
  }

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
  }
}

resource "aws_instance" "with_root_block_device_tags_and_ebs_block_device_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50

    tags = {
      RBDTagKey       = "rbd_tag_val"
      DefaultOverride = "rbd_tag_override"
    }
  }

  ebs_block_device {
    device_name = "my_data"
    volume_size = 50
    tags = {
      EBDTagKey       = "ebd_tag_val"
      DefaultOverride = "ebd_tag_override"
    }
  }
}
