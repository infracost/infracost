provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_security_group" "my_aws_security_group" {
}

resource "aws_mq_configuration" "my_aws_mq_configuration" {
  description    = "Example Configuration"
  name           = "example"
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}

resource "aws_mq_configuration" "my_aws_mq_configuration_rabbitmq" {
  description    = "Example Configuration"
  name           = "example"
  engine_type    = "RabbitMQ"
  engine_version = "3.8.11"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}

resource "aws_mq_broker" "my_aws_mq_broker_activemq_single_default" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.my_aws_mq_configuration.id
    revision = aws_mq_configuration.my_aws_mq_configuration.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.my_aws_security_group.id]
  deployment_mode    = "SINGLE_INSTANCE"

  user {
    username = "ExampleUser"
    password = "MindTheGappp"
  }
}

resource "aws_mq_broker" "my_aws_mq_broker_activemq_single_ebs" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.my_aws_mq_configuration.id
    revision = aws_mq_configuration.my_aws_mq_configuration.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.my_aws_security_group.id]
  deployment_mode    = "SINGLE_INSTANCE"
  storage_type       = "ebs"

  user {
    username = "ExampleUser"
    password = "MindTheGappp"
  }
}

resource "aws_mq_broker" "my_aws_mq_broker_activemq_standby_efs" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.my_aws_mq_configuration.id
    revision = aws_mq_configuration.my_aws_mq_configuration.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.m5.large"
  storage_type       = "efs"
  security_groups    = [aws_security_group.my_aws_security_group.id]
  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"

  user {
    username = "ExampleUser"
    password = "MindTheGappp"
  }
}

resource "aws_mq_broker" "my_aws_mq_broker_rabbitmq_single" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.my_aws_mq_configuration.id
    revision = aws_mq_configuration.my_aws_mq_configuration.latest_revision
  }

  engine_type        = "RabbitMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.m5.xlarge"
  security_groups    = [aws_security_group.my_aws_security_group.id]
  deployment_mode    = "SINGLE_INSTANCE"

  user {
    username = "ExampleUser"
    password = "MindTheGappp"
  }
}

resource "aws_mq_broker" "my_aws_mq_broker_rabbitmq_cluster" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.my_aws_mq_configuration_rabbitmq.id
    revision = aws_mq_configuration.my_aws_mq_configuration_rabbitmq.latest_revision
  }

  engine_type        = "RabbitMQ"
  engine_version     = "3.8.11"
  host_instance_type = "mq.m5.xlarge"
  security_groups    = [aws_security_group.my_aws_security_group.id]
  deployment_mode    = "cluster_multi_az"

  user {
    username = "ExampleUser"
    password = "MindTheGappp"
  }
}
