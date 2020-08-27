resource "aws_instance" "instance1" {
  ami           = var.aws_ami_id
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
  availability_zone = var.availability_zone_names[0]
  size              = 15
  type              = "standard"
}

resource "aws_ebs_volume" "io1" {
  availability_zone = var.availability_zone_names[0]
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
  image_id      = var.aws_ami_id
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
  image_id      = var.aws_ami_id
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

resource "aws_autoscaling_group" "mixed-instance-lt1" {
  desired_capacity   = 6
  max_size           = 10
  min_size           = 1

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.lt1.id
      }

      override {
        instance_type     = "t3.medium"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.large"
        weighted_capacity = "4"
      }
    }

    instances_distribution {
      on_demand_base_capacity = 1
      on_demand_percentage_above_base_capacity = 50
    }
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
  engine               = "postgres"
  instance_class       = "db.t2.small"
}

resource "aws_db_instance" "db_postgresql_iops" {
  allocated_storage    = 20
  engine               = "postgres"
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
  subnet_id     = var.aws_subnet_ids[0]
}

resource "aws_rds_cluster" "rds_cluster_mysql2" {
  cluster_identifier      = "aurora-cluster-demo"
  engine                  = "aurora-mysql"
  engine_version          = "5.7.mysql_aurora.2.03.2"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "bar"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}

resource "aws_rds_cluster" "rd_cluster_mulitmaster" {
  cluster_identifier   = "example"
  db_subnet_group_name = aws_db_subnet_group.example.name
  engine_mode          = "multimaster"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}