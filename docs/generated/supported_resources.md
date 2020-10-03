---
slug: supported_resources
title: Supported resources
---

Currently Infracost supports the following Terraform resources on AWS.

Support for the following is not currently included:
  * any costs that are not specified in the Terraform configuration, e.g. S3 storage costs, data out costs.
  * Any non On-Demand pricing, such as Reserved Instances.

| Terraform resource           | Notes |
| ---                          | ---   |
| `aws_alb` |  |
| `aws_autoscaling_group` |  See aws_instance<br />  |
| `aws_db_instance` |  |
| `aws_docdb_cluster_instance` |  |
| `aws_dynamodb_table` |  DAX is not yet supported.<br />  |
| `aws_ebs_snapshot` |  |
| `aws_ebs_snapshot_copy` |  |
| `aws_ebs_volume` |  |
| `aws_ecs_cluster` |  |
| `aws_ecs_service` |  Only supports Fargate on-demand.<br />  |
| `aws_ecs_task_definition` |  |
| `aws_elasticsearch_domain` |  |
| `aws_elb` |  |
| `aws_instance` |  Costs associated with non-standard Linux AMIs, such as Windows and RHEL are not supported.<br />  EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.<br />  If a root volume is not specified then an 8Gi gp2 volume is assumed.<br />  |
| `aws_lambda_function` |  Provisioned concurrency is not yet supported.<br />  |
| `aws_launch_configuration` |  |
| `aws_launch_template` |  |
| `aws_lb` |  |
| `aws_nat_gateway` |  |
| `aws_rds_cluster` |  |
| `aws_rds_cluster_instance` |  |


## The resource I want isn't supported

We're regularly adding support for new Terraform AWS resources - be sure to watch the repo for new [releases](https://github.com/infracost/infracost/releases)! We plan to add support for more cloud vendors ([GCP](https://github.com/infracost/infracost/issues/24), Azure) and other IaC tools (Pulumi) too.

You can help by:
1. [Creating an issue](https://github.com/infracost/infracost/issues/new) and mentioning the resource you need and a little about your use-case; we'll try to prioritize it depending on the community feedback.
2. [Contributing to Infracost](https://github.com/infracost/infracost#contributing), we're working on making it easier to add new resources. You can join our [Discord community](https://discord.gg/rXCTaH3) if you need help contributing.
