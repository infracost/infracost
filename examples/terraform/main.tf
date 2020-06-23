resource "aws_instance" "instance1" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdg"
    volume_type = "standard"
    volume_size = 20
  }

  ebs_block_device {
    device_name = "xvdh"
    volume_type = "sc1"
    volume_size = 30
  }

  ebs_block_device {
    device_name = "xvdi"
    volume_type = "io1"
    volume_size = 40
    iops        = 1000
  }
}

resource "aws_ebs_volume" "standard" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 15
  type              = "standard"
}

resource "aws_ebs_volume" "io1" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 10
  iops              = 500
}

resource "aws_ebs_snapshot" "standard" {
  volume_id = aws_ebs_volume.standard.id
}

resource "aws_ebs_snapshot_copy" "standard" {
  source_snapshot_id = aws_ebs_snapshot.standard.id
  source_region      = data.aws_region.current.name
}

resource "aws_launch_configuration" "lc1" {
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "t3.small"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }
}

resource "aws_autoscaling_group" "lc1" {
  launch_configuration = aws_launch_configuration.lc1.id
  desired_capacity     = 2
  max_size             = 2
  min_size             = 1
}

resource "aws_launch_template" "lt1" {
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "t3.small"

  block_device_mappings {
    device_name = "xvdf"
    ebs {
      volume_size = 10
    }
  }

  block_device_mappings {
    device_name = "xvfa"
    ebs {
      volume_size = 20
      volume_type = "io1"
      iops        = 200
    }
  }
}

resource "aws_autoscaling_group" "lt1" {
  desired_capacity = 1
  max_size         = 2
  min_size         = 0

  launch_template {
    id = aws_launch_template.lt1.id
  }
}

resource "aws_db_instance" "db_mysql" {
  allocated_storage    = 20
  storage_type         = "standard"
  engine               = "mysql"
  instance_class       = "db.t2.micro"
}

resource "aws_db_instance" "db_postgresql" {
  allocated_storage    = 20
  engine               = "PostgreSQL"
  instance_class       = "db.t2.small"
}

resource "aws_db_instance" "db_postgresql_iops" {
  allocated_storage    = 20
  engine               = "PostgreSQL"
  instance_class       = "db.t2.small"
  storage_type         = "io1"
  iops                 = 400
}

resource "aws_db_instance" "db_sqlserver" {
  allocated_storage    = 20
  engine               = "sqlserver-se"
  multi_az             = true
  instance_class       = "db.t3.xlarge"
}

resource "aws_elb" "elb1" {
  listener {
    instance_port     = 80
    instance_protocol = "HTTP"
    lb_port           = 80
    lb_protocol       = "HTTP"
  }
}

resource "aws_alb" "alb1" {
  load_balancer_type = "application"
}

resource "aws_lb" "nlb1" {
  load_balancer_type = "network"
}

resource "aws_eip" "nat1" {
  vpc = true
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat1.id
  subnet_id     = tolist(data.aws_subnet_ids.all.ids)[0]
}
