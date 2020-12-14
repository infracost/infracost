---
slug: supported_resources
title: Supported resources
---

Currently Infracost supports the following Terraform resources on AWS.

Support for the following is not currently included:
  * any costs that are not specified in the Terraform configuration, e.g. data out costs.
  * Any non On-Demand pricing, such as Reserved Instances.

| Terraform resource           | Notes |
| ---                          | ---   |
| `aws_alb` |  |
| `aws_api_gateway_rest_api` |  |
| `aws_api_gateway_stage` |  |
| `aws_apigatewayv2_api` |  |
| `aws_autoscaling_group` |  |
| `aws_cloudwatch_dashboard` |  |
| `aws_cloudwatch_log_group` |  |
| `aws_cloudwatch_metric_alarm` |  |
| `aws_db_instance` |  |
| `aws_dms_replication_instance` |  |
| `aws_docdb_cluster_instance` |  |
| `aws_dynamodb_table` |  DAX is not yet supported.<br />  |
| `aws_ebs_snapshot` |  |
| `aws_ebs_snapshot_copy` |  |
| `aws_ebs_volume` |  |
| `aws_ec2_traffic_mirror_session` |  |
| `aws_ecr_repository` |  |
| `aws_ecs_service` |  Only supports Fargate on-demand.<br />  |
| `aws_eip` |  |
| `aws_eks_cluster` |  |
| `aws_eks_fargate_profile` |  |
| `aws_eks_node_group` |  |
| `aws_elasticache_cluster` |  |
| `aws_elasticache_replication_group` |  |
| `aws_elasticsearch_domain` |  |
| `aws_elb` |  |
| `aws_instance` |  Costs associated with non-standard Linux AMIs, such as Windows and RHEL are not supported.<br />  EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.<br />  If a root volume is not specified then an 8Gi gp2 volume is assumed.<br />  |
| `aws_kms_external_key` |  |
| `aws_kms_key` |  |
| `aws_lambda_function` |  Provisioned concurrency is not yet supported.<br />  |
| `aws_lb` |  |
| `aws_lightsail_instance` |  |
| `aws_msk_cluster` |  |
| `aws_nat_gateway` |  |
| `aws_rds_cluster_instance` |  |
| `aws_route53_record` |  |
| `aws_route53_zone` |  |
| `aws_s3_bucket` |  S3 replication time control data transfer, and batch operations are not supported by Terraform.<br />  |
| `aws_s3_bucket_analytics_configuration` |  |
| `aws_s3_bucket_inventory` |  |
| `aws_sns_topic` |  |
| `aws_sns_topic_subscription` |  SMS and mobile push not yet supported.<br />  |
| `aws_sqs_queue` |  |
| `aws_vpc_endpoint` |  |
| `google_compute_address` |  |
| `google_compute_disk` |  |
| `google_compute_global_address` |  |
| `google_compute_image` |  |
| `google_compute_instance` |  Sustained use discounts are applied to monthly costs, but not to hourly costs.<br />  Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.<br />  Custom machine types are not supported.<br />  Sole-tenant VMs are not supported.<br />  |
| `google_compute_snapshot` |  |


## The resource I want isn't supported

We're regularly adding support for new Terraform AWS resources - be sure to watch the repo for new [releases](https://github.com/infracost/infracost/releases)! We plan to add support for more cloud vendors ([GCP](https://github.com/infracost/infracost/issues/24), Azure) and other IaC tools (Pulumi) too.

You can help by:
1. [Creating an issue](https://github.com/infracost/infracost/issues/new) and mentioning the resource you need and a little about your use-case; we'll try to prioritize it depending on the community feedback.
2. [Contributing to Infracost](https://github.com/infracost/infracost#contributing). You can join our [community Slack channel](https://www.infracost.io/community-chat) if you need help contributing.
