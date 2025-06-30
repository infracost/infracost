# type: ignore
import argparse

from aider.coders import Coder
from aider.coders.architect_coder import ArchitectCoder
from aider.coders.ask_coder import AskCoder
from aider.io import InputOutput
from aider.models import Model


def create_parser():
    parser = argparse.ArgumentParser(description="Coder script with dry-run option")
    parser.add_argument("--dry-run", action="store_true", help="Run in dry-run mode")
    return parser


args = create_parser().parse_args()

model = Model("claude-3-5-sonnet-20241022")
arch_model = Model("o1-preview-2024-09-12")

# Create a coder object
coder = Coder.create(
    main_model=model,
    io=InputOutput(
        yes=True,
        dry_run=args.dry_run,
    ),
)

ask = AskCoder.create(
    from_coder=coder,
)

arch = ArchitectCoder.create(
    from_coder=coder,
    main_model=arch_model,
)



files = [
"internal/providers/terraform/aws/acm_certificate_test.go",
"internal/providers/terraform/aws/acm_certificate.go",
"internal/providers/terraform/aws/acmpca_certificate_authority_test.go",
"internal/providers/terraform/aws/acmpca_certificate_authority.go",
"internal/providers/terraform/aws/api_gateway_rest_api_test.go",
"internal/providers/terraform/aws/api_gateway_rest_api.go",
"internal/providers/terraform/aws/api_gateway_stage_test.go",
# "internal/providers/terraform/aws/api_gateway_stage.go",
# "internal/providers/terraform/aws/apigatewayv2_api_test.go",
# "internal/providers/terraform/aws/apigatewayv2_api.go",
# "internal/providers/terraform/aws/app_autoscaling_target.go",
# "internal/providers/terraform/aws/autoscaling_group_test.go",
# "internal/providers/terraform/aws/autoscaling_group.go",
# "internal/providers/terraform/aws/aws_test.go",
# "internal/providers/terraform/aws/aws.go",
# "internal/providers/terraform/aws/backup_vault_test.go",
# "internal/providers/terraform/aws/backup_vault.go",
# "internal/providers/terraform/aws/cloudformation_stack_set_test.go",
# "internal/providers/terraform/aws/cloudformation_stack_set.go",
# "internal/providers/terraform/aws/cloudformation_stack_test.go",
# "internal/providers/terraform/aws/cloudformation_stack.go",
# "internal/providers/terraform/aws/cloudfront_distribution_test.go",
# "internal/providers/terraform/aws/cloudfront_distribution.go",
# "internal/providers/terraform/aws/cloudhsm_v2_hsm_test.go",
# "internal/providers/terraform/aws/cloudhsm_v2_hsm.go",
# "internal/providers/terraform/aws/cloudtrail_test.go",
# "internal/providers/terraform/aws/cloudtrail.go",
# "internal/providers/terraform/aws/cloudwatch_dashboard_test.go",
# "internal/providers/terraform/aws/cloudwatch_dashboard.go",
# "internal/providers/terraform/aws/cloudwatch_event_bus_test.go",
# "internal/providers/terraform/aws/cloudwatch_event_bus.go",
# "internal/providers/terraform/aws/cloudwatch_event_target.go",
# "internal/providers/terraform/aws/cloudwatch_log_group_test.go",
# "internal/providers/terraform/aws/cloudwatch_log_group.go",
# "internal/providers/terraform/aws/cloudwatch_metric_alarm_test.go",
# "internal/providers/terraform/aws/cloudwatch_metric_alarm.go",
# "internal/providers/terraform/aws/codebuild_project_test.go",
# "internal/providers/terraform/aws/codebuild_project.go",
# "internal/providers/terraform/aws/config_config_rule_test.go",
# "internal/providers/terraform/aws/config_config_rule.go",
# "internal/providers/terraform/aws/config_configuration_recorder_test.go",
# "internal/providers/terraform/aws/config_configuration_recorder.go",
# "internal/providers/terraform/aws/config_organization_custom_rule_test.go",
# "internal/providers/terraform/aws/config_organization_custom_rule.go",
# "internal/providers/terraform/aws/config_organization_managed_rule_test.go",
# "internal/providers/terraform/aws/config_organization_managed_rule.go",
# "internal/providers/terraform/aws/data_transfer_test.go",
# "internal/providers/terraform/aws/data_transfer.go",
# "internal/providers/terraform/aws/db_instance_test.go",
# "internal/providers/terraform/aws/db_instance.go",
# "internal/providers/terraform/aws/directory_service_directory_test.go",
# "internal/providers/terraform/aws/directory_service_directory.go",
# "internal/providers/terraform/aws/dms_test.go",
# "internal/providers/terraform/aws/dms.go",
# "internal/providers/terraform/aws/docdb_cluster_instance_test.go",
# "internal/providers/terraform/aws/docdb_cluster_instance.go",
# "internal/providers/terraform/aws/docdb_cluster_snapshot_test.go",
# "internal/providers/terraform/aws/docdb_cluster_snapshot.go",
# "internal/providers/terraform/aws/docdb_cluster_test.go",
# "internal/providers/terraform/aws/docdb_cluster.go",
# "internal/providers/terraform/aws/dx_connection_test.go",
# "internal/providers/terraform/aws/dx_connection.go",
# "internal/providers/terraform/aws/dx_gateway_association_test.go",
# "internal/providers/terraform/aws/dx_gateway_association.go",
# "internal/providers/terraform/aws/dynamodb_table_test.go",
# "internal/providers/terraform/aws/dynamodb_table.go",
# "internal/providers/terraform/aws/ebs_snapshot_copy_test.go",
# "internal/providers/terraform/aws/ebs_snapshot_copy.go",
# "internal/providers/terraform/aws/ebs_snapshot_test.go",
# "internal/providers/terraform/aws/ebs_snapshot.go",
# "internal/providers/terraform/aws/ebs_volume_test.go",
# "internal/providers/terraform/aws/ebs_volume.go",
# "internal/providers/terraform/aws/ec2_client_vpn_endpoint_test.go",
# "internal/providers/terraform/aws/ec2_client_vpn_endpoint.go",
# "internal/providers/terraform/aws/ec2_client_vpn_network_association_test.go",
# "internal/providers/terraform/aws/ec2_client_vpn_network_association.go",
# "internal/providers/terraform/aws/ec2_host_test.go",
# "internal/providers/terraform/aws/ec2_host.go",
# "internal/providers/terraform/aws/ec2_traffic_mirror_session_test.go",
# "internal/providers/terraform/aws/ec2_traffic_mirror_session.go",
# "internal/providers/terraform/aws/ec2_transit_gateway_peering_attachment_test.go",
# "internal/providers/terraform/aws/ec2_transit_gateway_peering_attachment.go",
# "internal/providers/terraform/aws/ec2_transit_gateway_vpc_attachment_test.go",
# "internal/providers/terraform/aws/ec2_transit_gateway_vpc_attachment.go",
# "internal/providers/terraform/aws/ecr_lifecycle_policy.go",
# "internal/providers/terraform/aws/ecr_repository_test.go",
# "internal/providers/terraform/aws/ecr_repository.go",
# "internal/providers/terraform/aws/ecs_cluster_capacity_providers.go",
# "internal/providers/terraform/aws/ecs_cluster.go",
# "internal/providers/terraform/aws/ecs_service_internal_test.go",
# "internal/providers/terraform/aws/ecs_service_test.go",
# "internal/providers/terraform/aws/ecs_service.go",
# "internal/providers/terraform/aws/ecs_task_definition.go",
# "internal/providers/terraform/aws/ecs_task_set.go",
# "internal/providers/terraform/aws/efs_file_system_test.go",
# "internal/providers/terraform/aws/efs_file_system.go",
# "internal/providers/terraform/aws/eip_association.go",
# "internal/providers/terraform/aws/eip_test.go",
# "internal/providers/terraform/aws/eip.go",
# "internal/providers/terraform/aws/eks_cluster_test.go",
# "internal/providers/terraform/aws/eks_cluster.go",
# "internal/providers/terraform/aws/eks_fargate_profile_test.go",
# "internal/providers/terraform/aws/eks_fargate_profile.go",
# "internal/providers/terraform/aws/eks_node_group_test.go",
# "internal/providers/terraform/aws/eks_node_group.go",
# "internal/providers/terraform/aws/elastic_beanstalk_environment_test.go",
# "internal/providers/terraform/aws/elastic_beanstalk_environment.go",
# "internal/providers/terraform/aws/elasticache_cluster_test.go",
# "internal/providers/terraform/aws/elasticache_cluster.go",
# "internal/providers/terraform/aws/elasticache_replication_group_test.go",
# "internal/providers/terraform/aws/elasticache_replication_group.go",
# "internal/providers/terraform/aws/elasticsearch_domain_test.go",
# "internal/providers/terraform/aws/elasticsearch_domain.go",
# "internal/providers/terraform/aws/elb_test.go",
# "internal/providers/terraform/aws/elb.go",
# "internal/providers/terraform/aws/flow_log.go",
# "internal/providers/terraform/aws/fsx_openzfs_file_system_test.go",
# "internal/providers/terraform/aws/fsx_openzfs_file_system.go",
# "internal/providers/terraform/aws/fsx_windows_file_system_test.go",
# "internal/providers/terraform/aws/fsx_windows_file_system.go",
# "internal/providers/terraform/aws/global_accelerator_endpoint_group_test.go",
# "internal/providers/terraform/aws/global_accelerator_endpoint_group.go",
# "internal/providers/terraform/aws/global_accelerator_test.go",
# "internal/providers/terraform/aws/global_accelerator.go",
# "internal/providers/terraform/aws/glue_catalog_database_test.go",
# "internal/providers/terraform/aws/glue_catalog_database.go",
# "internal/providers/terraform/aws/glue_crawler_test.go",
# "internal/providers/terraform/aws/glue_crawler.go",
# "internal/providers/terraform/aws/glue_job_test.go",
# "internal/providers/terraform/aws/glue_job.go",
# "internal/providers/terraform/aws/instance_test.go",
# "internal/providers/terraform/aws/instance.go",
# "internal/providers/terraform/aws/kinesis_firehose_delivery_stream_test.go",
# "internal/providers/terraform/aws/kinesis_firehose_delivery_stream.go",
# "internal/providers/terraform/aws/kinesis_stream_test.go",
# "internal/providers/terraform/aws/kinesis_stream.go",
# "internal/providers/terraform/aws/kinesisanalytics_application_test.go",
# "internal/providers/terraform/aws/kinesisanalytics_application.go",
# "internal/providers/terraform/aws/kinesisanalyticsv2_application_snapshot_test.go",
# "internal/providers/terraform/aws/kinesisanalyticsv2_application_snapshot.go",
# "internal/providers/terraform/aws/kinesisanalyticsv2_application_test.go",
# "internal/providers/terraform/aws/kinesisanalyticsv2_application.go",
# "internal/providers/terraform/aws/kms_external_key_test.go",
# "internal/providers/terraform/aws/kms_external_key.go",
# "internal/providers/terraform/aws/kms_key_test.go",
# "internal/providers/terraform/aws/kms_key.go",
# "internal/providers/terraform/aws/lambda_function_test.go",
# "internal/providers/terraform/aws/lambda_function.go",
# "internal/providers/terraform/aws/lambda_provisioned_concurrency_config_test.go",
# "internal/providers/terraform/aws/lambda_provisioned_concurrency_config.go",
# "internal/providers/terraform/aws/lb_test.go",
# "internal/providers/terraform/aws/lb.go",
# "internal/providers/terraform/aws/lightsail_instance_test.go",
# "internal/providers/terraform/aws/lightsail_instance.go",
# "internal/providers/terraform/aws/mq_broker_test.go",
# "internal/providers/terraform/aws/mq_broker.go",
# "internal/providers/terraform/aws/msk_cluster_test.go",
# "internal/providers/terraform/aws/msk_cluster.go",
# "internal/providers/terraform/aws/mwaa_environment_test.go",
# "internal/providers/terraform/aws/mwaa_environment.go",
# "internal/providers/terraform/aws/nat_gateway_test.go",
# "internal/providers/terraform/aws/nat_gateway.go",
# "internal/providers/terraform/aws/neptune_cluster_instance_test.go",
# "internal/providers/terraform/aws/neptune_cluster_instance.go",
# "internal/providers/terraform/aws/neptune_cluster_snapshot_test.go",
# "internal/providers/terraform/aws/neptune_cluster_snapshot.go",
# "internal/providers/terraform/aws/neptune_cluster_test.go",
# "internal/providers/terraform/aws/neptune_cluster.go",
# "internal/providers/terraform/aws/networkfirewall_firewall_test.go",
# "internal/providers/terraform/aws/networkfirewall_firewall.go",
# "internal/providers/terraform/aws/opensearch_domain_test.go",
# "internal/providers/terraform/aws/opensearch_domain.go",
# "internal/providers/terraform/aws/pipes_pipe.go",
# "internal/providers/terraform/aws/rds_cluster_instance_test.go",
# "internal/providers/terraform/aws/rds_cluster_instance.go",
# "internal/providers/terraform/aws/rds_cluster_test.go",
# "internal/providers/terraform/aws/rds_cluster.go",
# "internal/providers/terraform/aws/redshift_cluster_test.go",
# "internal/providers/terraform/aws/redshift_cluster.go",
# "internal/providers/terraform/aws/registry.go",
# "internal/providers/terraform/aws/route53_health_check_test.go",
# "internal/providers/terraform/aws/route53_health_check.go",
# "internal/providers/terraform/aws/route53_record_test.go",
# "internal/providers/terraform/aws/route53_record.go",
# "internal/providers/terraform/aws/route53_resolver_endpoint_test.go",
# "internal/providers/terraform/aws/route53_resolver_endpoint.go",
# "internal/providers/terraform/aws/route53_zone_test.go",
# "internal/providers/terraform/aws/route53_zone.go",
# "internal/providers/terraform/aws/s3_bucket_analytics_configuration_test.go",
# "internal/providers/terraform/aws/s3_bucket_analytics_configuration.go",
# "internal/providers/terraform/aws/s3_bucket_intelligent_tiering_configuration.go",
# "internal/providers/terraform/aws/s3_bucket_inventory_test.go",
# "internal/providers/terraform/aws/s3_bucket_inventory.go",
# "internal/providers/terraform/aws/s3_bucket_lifecycle_configuration.go",
# "internal/providers/terraform/aws/s3_bucket_test.go",
# "internal/providers/terraform/aws/s3_bucket_versioning.go",
# "internal/providers/terraform/aws/s3_bucket.go",
# "internal/providers/terraform/aws/scheduler_schedule.go",
# "internal/providers/terraform/aws/secretsmanager_secret_test.go",
# "internal/providers/terraform/aws/secretsmanager_secret.go",
# "internal/providers/terraform/aws/sfn_state_machine_test.go",
# "internal/providers/terraform/aws/sfn_state_machine.go",
# "internal/providers/terraform/aws/sns_topic_subscription_test.go",
# "internal/providers/terraform/aws/sns_topic_subscription.go",
# "internal/providers/terraform/aws/sns_topic_test.go",
# "internal/providers/terraform/aws/sns_topic.go",
# "internal/providers/terraform/aws/spot_instance_request_test.go",
# "internal/providers/terraform/aws/spot_instance_request.go",
# "internal/providers/terraform/aws/sqs_queue_test.go",
# "internal/providers/terraform/aws/sqs_queue.go",
# "internal/providers/terraform/aws/ssm_activation_test.go",
# "internal/providers/terraform/aws/ssm_activation.go",
# "internal/providers/terraform/aws/ssm_parameter_test.go",
# "internal/providers/terraform/aws/ssm_parameter.go",
# "internal/providers/terraform/aws/subnet.go",
# "internal/providers/terraform/aws/tags.go",
# "internal/providers/terraform/aws/transfer_server_test.go",
# "internal/providers/terraform/aws/transfer_server.go",
# "internal/providers/terraform/aws/util.go",
# "internal/providers/terraform/aws/vpc_endpoint_test.go",
# "internal/providers/terraform/aws/vpc_endpoint.go",
# "internal/providers/terraform/aws/vpc.go",
# "internal/providers/terraform/aws/vpn_connection_test.go",
# "internal/providers/terraform/aws/vpn_connection.go",
# "internal/providers/terraform/aws/waf_web_acl_test.go",
# "internal/providers/terraform/aws/waf_web_acl.go",
# "internal/providers/terraform/aws/wafv2_web_acl_test.go",
# "internal/providers/terraform/aws/wafv2_web_acl.go",
]

conventions_file = "_CONVENTIONS.xml"

print(f"/read {conventions_file}")
coder.run(f"/read {conventions_file}")

for file in files:
    print(f"/add {file}")
    coder.run(f"/add {file}")

    prompt = f"""
    please read the conventions for our coding standards in the
    {conventions_file} file and apply them to the file {file}.
    Please follow the directions in the {conventions_file} to make changes to
    the file {file}. Follow the instructions in the {conventions_file} carefully and pay attention to the examples and correct all instances of the issues in the file {file}.
    """

    print("")
    print("")

    print(prompt)
    coder.run(prompt)
    
    print("")
    print("")

    print("/lint")
    coder.run("/lint")
    print("/lint")
    coder.run("/lint")

    print(f"/drop {file}")
    coder.run(f"/drop {file}")