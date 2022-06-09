package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	// GetAPIGatewayRestAPIRegistryItem(),
	// GetAPIGatewayStageRegistryItem(),
	// GetAPIGatewayv2ApiRegistryItem(),
	// GetAutoscalingGroupRegistryItem(),
	getACMCertificate(),
	// GetACMPCACertificateAuthorityRegistryItem(),
	// GetCloudfrontDistributionRegistryItem(),
	// GetCloudwatchDashboardRegistryItem(),
	// GetCloudwatchEventBusItem(),
	// GetCloudwatchLogGroupItem(),
	// GetCloudwatchMetricAlarmRegistryItem(),
	// GetCodebuildProjectRegistryItem(),
	// GetConfigRuleItem(),
	// GetConfigurationRecorderItem(),
	// GetConfigOrganizationCustomRuleItem(),
	// GetConfigOrganizationManagedRuleItem(),
	// getDataTransferRegistryItem(),
	getDBInstanceRegistryItem(),
	// GetDMSRegistryItem(),
	// GetDocDBClusterInstanceRegistryItem(),
	// GetDocDBClusterRegistryItem(),
	// GetDocDBClusterSnapshotRegistryItem(),
	// GetDXConnectionRegistryItem(),
	// GetDXGatewayAssociationRegistryItem(),
	getDynamoDBTableRegistryItem(),
	getEBSSnapshotCopyRegistryItem(),
	getEBSSnapshotRegistryItem(),
	getEBSVolumeRegistryItem(),
	// GetEC2ClientVPNEndpointRegistryItem(),
	// GetEC2ClientVPNNetworkAssociationRegistryItem(),
	// GetEC2TrafficMirroSessionRegistryItem(),
	// GetEC2TransitGatewayPeeringAttachmentRegistryItem(),
	// GetEC2TransitGatewayVpcAttachmentRegistryItem(),
	// GetECRRegistryItem(),
	// GetECSServiceRegistryItem(),
	// GetEFSFileSystemRegistryItem(),
	getEIPRegistryItem(),
	// GetElastiCacheClusterItem(),
	// GetElastiCacheReplicationGroupItem(),
	// GetElasticsearchDomainRegistryItem(),
	// GetELBRegistryItem(),
	// GetFSXWindowsFSRegistryItem(),
	getInstanceRegistryItem(),
	// GetLambdaFunctionRegistryItem(),
	// GetLBRegistryItem(),
	// GetLightsailInstanceRegistryItem(),
	// GetMSKClusterRegistryItem(),
	// GetALBRegistryItem(),
	// GetMQBrokerRegistryItem(),
	getNATGatewayRegistryItem(),
	getRDSClusterRegistryItem(),
	getRDSClusterInstanceRegistryItem(),
	// GetRedshiftClusterRegistryItem(),
	// GetRoute53HealthCheck(),
	// GetRoute53ResolverEndpointRegistryItem(),
	// GetRoute53RecordRegistryItem(),
	// GetRoute53ZoneRegistryItem(),
	getS3BucketRegistryItem(),
	// GetS3BucketAnalyticsConfigurationRegistryItem(),
	// GetS3BucketInventoryRegistryItem(),
	// GetSecretsManagerSecret(),
	// GetSSMActivationRegistryItem(),
	// GetSSMParameterRegistryItem(),
	// GetSNSTopicRegistryItem(),
	// GetSNSTopicSubscriptionRegistryItem(),
	// GetSQSQueueRegistryItem(),
	getNewEKSNodeGroupItem(),
	//getNewEKSFargateProfileItem(),
	getNewEKSClusterItem(),
	// GetNewKMSKeyRegistryItem(),
	// GetNewKMSExternalKeyRegistryItem(),
	// GetVPNConnectionRegistryItem(),
	// GetVpcEndpointRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources = []string{
	// AWS Application Auto Scaling
	"aws_appautoscaling_policy",
	"aws_appautoscaling_scheduled_action",

	// AWS Certificate Manager
	"aws_acm_certificate_validation",

	// AWS API Gateway Rest APIs
	"aws_api_gateway_account",
	"aws_api_gateway_api_key",
	"aws_api_gateway_authorizer",
	"aws_api_gateway_base_path_mapping",
	"aws_api_gateway_client_certificate",
	"aws_api_gateway_deployment",
	"aws_api_gateway_documentation_part",
	"aws_api_gateway_documentation_version",
	"aws_api_gateway_domain_name",
	"aws_api_gateway_response",
	"aws_api_gateway_integration",
	"aws_api_gateway_method",
	"aws_api_gateway_method_response",
	"aws_api_gateway_method_settings",
	"aws_api_gateway_model",
	"aws_api_gateway_request_validator",
	"aws_api_gateway_resource",
	"aws_api_gateway_usage_plan",
	"aws_api_gateway_usage_plan_key",
	"aws_api_gateway_vpc_link",

	// AWS API Gateway v2 HTTP & Websocket API.
	"aws_apigatewayv2_api_mapping",
	"aws_apigatewayv2_authorizer",
	"aws_apigatewayv2_deployment",
	"aws_apigatewayv2_domain_name",
	"aws_apigatewayv2_integration",
	"aws_apigatewayv2_integration_response",
	"aws_apigatewayv2_model",
	"aws_apigatewayv2_route",
	"aws_apigatewayv2_route_response",
	"aws_apigatewayv2_stage",
	"aws_apigatewayv2_vpc_link",

	// AWS Backup
	"aws_backup_global_settings",
	"aws_backup_plan",
	"aws_backup_region_settings",
	"aws_backup_selection",
	"aws_backup_vault_notifications",
	"aws_backup_vault_policy",

	// AWS DX Transit.
	"aws_dx_bgp_peer",
	"aws_dx_gateway",
	"aws_dx_gateway_association_proposal",
	"aws_dx_hosted_private_virtual_interface",
	"aws_dx_hosted_private_virtual_interface_accepter",
	"aws_dx_hosted_public_virtual_interface",
	"aws_dx_hosted_public_virtual_interface_accepter",
	"aws_dx_hosted_transit_virtual_interface",
	"aws_dx_hosted_transit_virtual_interface_accepter",
	"aws_dx_lag",
	"aws_dx_private_virtual_interface",
	"aws_dx_public_virtual_interface",
	"aws_dx_transit_virtual_interface",

	// AWS Cloudformation
	"aws_cloudformation_stack_set_instance",
	"aws_cloudformation_type",

	// AWS Cloudfront
	"aws_cloudfront_origin_access_identity",
	"aws_cloudfront_public_key",

	// AWS Cloudwatch
	"aws_cloudwatch_log_destination",
	"aws_cloudwatch_log_destination_policy",
	"aws_cloudwatch_log_metric_filter",
	"aws_cloudwatch_log_resource_policy",
	"aws_cloudwatch_log_stream",
	"aws_cloudwatch_log_subscription_filter",

	// AWS EventBridge
	"aws_cloudwatch_event_permission",
	"aws_cloudwatch_event_rule",
	"aws_cloudwatch_event_target",

	// AWS CodeBuild
	"aws_codebuild_report_group",
	"aws_codebuild_source_credential",
	"aws_codebuild_webhook",

	// AWS Config
	"aws_config_aggregate_authorization",
	"aws_config_configuration_aggregator",
	"aws_config_configuration_recorder_status",
	"aws_config_delivery_channel",
	"aws_config_remediation_configuration",

	// AWS ECR
	"aws_ecr_lifecycle_policy",
	"aws_ecr_repository_policy",

	// AWS Elastic Beanstalk
	"aws_elastic_beanstalk_application",

	// AWS Elastic Container Service
	"aws_ecs_account_setting_default",

	// AWS Elastic File System
	"aws_efs_access_point",
	"aws_efs_file_system_policy",
	"aws_efs_mount_target",

	// AWS Elastic Load Balancing
	"aws_alb_listener",
	"aws_alb_listener_certificate",
	"aws_alb_listener_rule",
	"aws_alb_target_group",
	"aws_alb_target_group_attachment",
	"aws_lb_listener",
	"aws_lb_listener_certificate",
	"aws_lb_listener_rule",
	"aws_lb_target_group",
	"aws_lb_target_group_attachment",
	"aws_app_cookie_stickiness_policy",
	"aws_elb_attachment",
	"aws_lb_cookie_stickiness_policy",
	"aws_lb_ssl_negotiation_policy",
	"aws_load_balancer_backend_server_policy",
	"aws_load_balancer_listener_policy",
	"aws_load_balancer_policy",

	// AWS Elasticache
	"aws_elasticache_parameter_group",
	"aws_elasticache_security_group",
	"aws_elasticache_subnet_group",

	// AWS Glue
	"aws_glue_catalog_table",
	"aws_glue_classifier",
	"aws_glue_connection",
	"aws_glue_data_catalog_encryption_settings",
	"aws_glue_partition",
	"aws_glue_partition_index",
	"aws_glue_registry",
	"aws_glue_resource_policy",
	"aws_glue_schema",
	"aws_glue_security_configuration",
	"aws_glue_trigger",
	"aws_glue_user_defined_function",
	"aws_glue_workflow",

	// AWS IAM aws_iam_* resources
	"aws_iam_access_key",
	"aws_iam_account_alias",
	"aws_iam_account_alias",
	"aws_iam_account_password_policy",
	"aws_iam_group",
	"aws_iam_group",
	"aws_iam_group_membership",
	"aws_iam_group_policy",
	"aws_iam_group_policy_attachment",
	"aws_iam_instance_profile",
	"aws_iam_instance_profile",
	"aws_iam_openid_connect_provider",
	"aws_iam_policy",
	"aws_iam_policy",
	"aws_iam_policy_attachment",
	"aws_iam_role",
	"aws_iam_role",
	"aws_iam_role_policy",
	"aws_iam_role_policy_attachment",
	"aws_iam_rolepolicyattachment",
	"aws_iam_saml_provider",
	"aws_iam_server_certificate",
	"aws_iam_server_certificate",
	"aws_iam_service_linked_role",
	"aws_iam_user",
	"aws_iam_user",
	"aws_iam_user_group_membership",
	"aws_iam_user_login_profile",
	"aws_iam_user_policy",
	"aws_iam_user_policy_attachment",
	"aws_iam_user_ssh_key",

	// AWS IOT
	"aws_iot_policy",

	// AWS KMS
	"aws_kms_alias",
	"aws_kms_ciphertext",
	"aws_kms_grant",

	// AWS Lambda
	"aws_lambda_alias",
	"aws_lambda_code_signing_config",
	"aws_lambda_event_source_mapping",
	"aws_lambda_function_event_invoke_config",
	"aws_lambda_layer_version",
	"aws_lambda_layer_version_permission",
	"aws_lambda_permission",

	// AWS Neptune
	"aws_neptune_cluster_parameter_group",
	"aws_neptune_event_subscription",
	"aws_neptune_parameter_group",
	"aws_neptune_subnet_group",

	// AWS Network Firewall
	"aws_networkfirewall_rule_group",
	"aws_networkfirewall_firewall_policy",
	"aws_networkfirewall_logging_configuration",

	// AWS Others
	"aws_db_instance_role_association",
	"aws_db_option_group",
	"aws_db_parameter_group",
	"aws_db_subnet_group",
	"aws_dms_replication_subnet_group",
	"aws_dms_replication_task",
	"aws_docdb_cluster_parameter_group",
	"aws_docdb_subnet_group",
	"aws_dynamodb_table_item",
	"aws_ebs_encryption_by_default",
	"aws_ebs_default_kms_key",
	"aws_ecs_task_definition",
	"aws_elasticsearch_domain_policy",
	"aws_key_pair",
	"aws_launch_configuration",
	"aws_launch_template",
	"aws_lightsail_domain",
	"aws_lightsail_key_pair",
	"aws_lightsail_static_ip",
	"aws_lightsail_static_ip_attachment",
	"aws_mq_configuration",
	"aws_msk_configuration",
	"aws_rds_cluster_endpoint",
	"aws_rds_cluster_parameter_group",
	"aws_resourcegroups_group",
	"aws_route53_resolver_dnssec_config",
	"aws_route53_resolver_query_log_config",
	"aws_route53_resolver_query_log_config_association",
	"aws_route53_resolver_rule",
	"aws_route53_resolver_rule_association",
	"aws_route53_zone_association",
	"aws_sqs_queue_policy",
	"aws_volume_attachment",

	// AWS RAM
	"aws_ram_principal_association",
	"aws_ram_resource_association",
	"aws_ram_resource_share",
	"aws_ram_resource_share_accepter",

	// AWS S3
	"aws_s3_access_point",
	"aws_s3_account_public_access_block",
	"aws_s3_bucket_acl",
	"aws_s3_bucket_metric",
	"aws_s3_bucket_notification",
	"aws_s3_bucket_object", // Costs are shown at the bucket level
	"aws_s3_bucket_ownership_controls",
	"aws_s3_bucket_policy",
	"aws_s3_bucket_public_access_block",

	// AWS Secrets Manager
	"aws_secretsmanager_secret_policy",
	"aws_secretsmanager_secret_rotation",
	"aws_secretsmanager_secret_version",

	// AWS Service Discovery Service
	"aws_service_discovery_service",

	// AWS SES
	"aws_ses_domain_dkim",
	"aws_ses_domain_identity",

	// AWS SNS
	"aws_sns_platform_application",
	"aws_sns_sms_preferences",
	"aws_sns_topic_policy",

	// AWS SSM
	"aws_ssm_association",
	"aws_ssm_maintenance_window",
	"aws_ssm_maintenance_window_target",
	"aws_ssm_maintenance_window_task",
	"aws_ssm_patch_baseline",
	"aws_ssm_patch_group",
	"aws_ssm_resource_data_sync",

	// AWS Transfer Family
	"aws_transfer_access",
	"aws_transfer_ssh_key",
	"aws_transfer_user",

	// AWS VPC
	"aws_customer_gateway",
	"aws_default_network_acl",
	"aws_default_route_table",
	"aws_default_security_group",
	"aws_default_subnet",
	"aws_default_vpc",
	"aws_default_vpc_dhcp_options",
	"aws_ec2_client_vpn_authorization_rule",
	"aws_ec2_client_vpn_route",
	"aws_ec2_tag",
	"aws_ec2_traffic_mirror_filter",
	"aws_ec2_traffic_mirror_filter_rule",
	"aws_ec2_traffic_mirror_target",
	"aws_ec2_transit_gateway",
	"aws_ec2_transit_gateway_peering_attachment_accepter",
	"aws_ec2_transit_gateway_route",
	"aws_ec2_transit_gateway_route_table",
	"aws_ec2_transit_gateway_route_table_association",
	"aws_ec2_transit_gateway_route_table_propagation",
	"aws_ec2_transit_gateway_vpc_attachment_accepter",
	"aws_egress_only_internet_gateway",
	"aws_flow_log",
	"aws_internet_gateway",
	"aws_ec2_internetgateway",
	"aws_main_route_table_association",
	"aws_network_acl",
	"aws_network_acl_rule",
	"aws_network_interface",
	"aws_network_interface_attachment",
	"aws_network_interface_sg_attachment",
	"aws_route",
	"aws_ec2_route",
	"aws_route_table",
	"aws_ec2_routetable",
	"aws_route_table_association",
	"aws_ec2_routetableassociation",
	"aws_security_group",
	"aws_ec2_securitygroup",
	"aws_security_group_rule",
	"aws_subnet",
	"aws_ec2_subnet",
	"aws_vpc",
	"aws_ec2_vpc",
	"aws_vpc_dhcp_options",
	"aws_vpc_dhcp_options_association",
	"aws_vpc_endpoint_connection_notification",
	"aws_vpc_endpoint_route_table_association",
	"aws_vpc_endpoint_service",
	"aws_vpc_endpoint_service_allowed_principal",
	"aws_vpc_endpoint_subnet_association",
	"aws_vpc_ipv4_cidr_block_association",
	"aws_vpc_peering_connection",
	"aws_vpc_peering_connection_accepter",
	"aws_vpc_peering_connection_options",
	"aws_vpn_connection_route",
	"aws_vpn_gateway",
	"aws_vpn_gateway_attachment",
	"aws_vpn_gateway_route_propagation",

	// EIP Association
	"aws_ec2_eipassociation",

	// Hashicorp
	"null_resource",
	"local_file",
	"template_dir",
	"random_id",
	"random_integer",
	"random_password",
	"random_pet",
	"random_shuffle",
	"random_string",
	"random_uuid",
	"tls_locally_signed_cert",
	"tls_private_key",
	"tls_self_signed_cert",
	"time_offset",
	"time_rotating",
	"time_sleep",
	"time_static",

	// WAF
	"aws_wafv2_rule_group",
	"aws_wafv2_ip_set",
	"aws_wafv2_regex_pattern_set",
	"aws_wafv2_web_acl_association",
	"aws_wafv2_web_acl_logging_configuration",
	"aws_waf_byte_match_set",
	"aws_waf_geo_match_set",
	"aws_waf_ipset",
	"aws_waf_regex_match_set",
	"aws_waf_regex_pattern_set",
	"aws_waf_size_constraint_set",
	"aws_waf_sql_injection_match_set",
	"aws_waf_xss_match_set",
	"aws_waf_rule",
	"aws_waf_rate_based_rule",
	"aws_waf_rule_group",
}

var UsageOnlyResources = []string{
	"aws_data_transfer",
}
