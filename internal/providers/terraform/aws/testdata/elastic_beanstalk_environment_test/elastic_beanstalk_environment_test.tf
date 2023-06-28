provider "aws" {
  region                      = "eu-west-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for ElasticBeanstalkEnvironment below

resource "aws_elastic_beanstalk_application" "my_eb_application" {
  name        = "eb_application"
  description = "description"
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment_with_usage" {
  name                = "eb_environment_with_usage"
  application         = aws_elastic_beanstalk_application.my_eb_application.name
  solution_stack_name = "64bit Amazon Linux 2 v3.4.11 running Docker"

  setting {
    namespace = "aws:ec2:instances"
    name      = "InstanceTypes"
    value     = "c4.large"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = 4
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = 6
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "LoadBalancerType"
    value     = "classic"
  }

  setting {
    namespace = "aws:elasticbeanstalk:cloudwatch:logs"
    name      = "StreamLogs"
    value     = true
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "RootVolumeType"
    value     = "io1"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "RootVolumeIOPS"
    value     = 300
  }
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment" {
  name                = "eb_environment"
  application         = aws_elastic_beanstalk_application.my_eb_application.name
  solution_stack_name = "64bit Amazon Linux 2 v3.4.11 running Docker"

  setting {
    namespace = "aws:ec2:instances"
    name      = "InstanceTypes"
    value     = "t3.small"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = 1
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = 4
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "LoadBalancerType"
    value     = "network"
  }
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment_with_rds" {
  name                = "eb_environment_with_rds"
  application         = aws_elastic_beanstalk_application.my_eb_application.name
  solution_stack_name = "64bit Amazon Linux 2 v3.4.11 running Docker"

  setting {
    namespace = "aws:ec2:instances"
    name      = "InstanceTypes"
    value     = "t3.small"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = 2
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = 4
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "LoadBalancerType"
    value     = "network"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "DBInstanceClass"
    value     = "db.m6g.xlarge"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "DBEngine"
    value     = "postgres"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "MultiAZDatabase"
    value     = true
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "DBAllocatedStorage"
    value     = 100
  }
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment_with_rds_no_usage" {
  name                = "eb_environment_with_rds_no_usage"
  application         = aws_elastic_beanstalk_application.my_eb_application.name
  solution_stack_name = "64bit Amazon Linux 2 v3.4.11 running Docker"

  setting {
    namespace = "aws:ec2:instances"
    name      = "InstanceTypes"
    value     = "t3.large"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = 2
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = 4
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "LoadBalancerType"
    value     = "network"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "DBInstanceClass"
    value     = "db.m6g.xlarge"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "DBEngine"
    value     = "postgres"
  }

  setting {
    namespace = "aws:rds:dbinstance"
    name      = "MultiAZDatabase"
    value     = true
  }

  setting {
    namespace = "aws:elasticbeanstalk:cloudwatch:logs"
    name      = "StreamLogs"
    value     = true
  }
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment_asg_instance_type" {
  name        = "eb_environment_asg"
  application = aws_elastic_beanstalk_application.my_eb_application.name

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "t3a.large"
  }
}

resource "aws_elastic_beanstalk_environment" "my_eb_environment_asg_instance_types" {
  name        = "eb_environment_asg"
  application = aws_elastic_beanstalk_application.my_eb_application.name

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "t3a.large"
  }

  // This should override the `InstanceType` setting
  setting {
    namespace = "aws:ec2:instances"
    name      = "InstanceTypes"
    value     = "t3a.xlarge"
  }
}
