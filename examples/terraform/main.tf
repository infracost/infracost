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

resource "aws_rds_cluster_instance" "cluster_instances" {
  count              = 2
  identifier         = "aurora-cluster-demo-${count.index}"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_dynamodb_table" "dynamodb-table" {
  name           = "GameScores"
  billing_mode   = "PROVISIONED"
  read_capacity  = 30
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  replica {
    region_name = "us-east-2"
  }

  replica {
    region_name = "us-west-1"
  }
}

resource "aws_ecs_cluster" "ecs1" {
  name               = "ecs1"
  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_task_definition" "ecs_task1" {
  requires_compatibilities = ["FARGATE"]
  family                   = "ecs_task1"
  memory                   = "2 GB"
  cpu                      = "1 vCPU"

  inference_accelerator {
    device_name = "device1"
    device_type = "eia2.medium"
  }

  container_definitions = <<TASK_DEFINITION
    [
        {
            "command": ["sleep", "10"],
            "entryPoint": ["/"],
            "essential": true,
            "image": "alpine",
            "name": "alpine",
            "network_mode": "none"
        }
    ]
  TASK_DEFINITION
}

resource "aws_ecs_service" "ecs_fargate1" {
  name            = "ecs_fargate1"
  launch_type     = "FARGATE"
  cluster         = aws_ecs_cluster.ecs1.id
  task_definition = aws_ecs_task_definition.ecs_task1.arn
  desired_count   = 2
}
