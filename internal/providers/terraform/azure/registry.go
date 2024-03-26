package azure

import "github.com/infracost/infracost/internal/schema"

// ResourceRegistry grouped alphabetically
var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getActiveDirectoryDomainServiceRegistryItem(),
	getActiveDirectoryDomainServiceReplicaSetRegistryItem(),
	getAPIManagementRegistryItem(),
	getApplicationGatewayRegistryItem(),
	getAppServiceEnvironmentRegistryItem(),
	GetAzureRMAppIntegrationServiceEnvironmentRegistryItem(),
	getFunctionAppRegistryItem(),
	GetAzureRMAppNATGatewayRegistryItem(),
	getAppServiceCertificateBindingRegistryItem(),
	getAppServiceCertificateOrderRegistryItem(),
	getAppServiceCustomHostnameBindingRegistryItem(),
	getAppServicePlanRegistryItem(),
	getApplicationInsightsWebTestRegistryItem(),
	getApplicationInsightsRegistryItem(),
	getAutomationAccountRegistryItem(),
	getAutomationDSCConfigurationRegistryItem(),
	getAutomationDSCNodeConfigurationRegistryItem(),
	getAutomationJobScheduleRegistryItem(),
	getBastionHostRegistryItem(),
	GetAzureRMCDNEndpointRegistryItem(),
	getContainerRegistryRegistryItem(),
	getCosmosDBAccountRegistryItem(),
	GetAzureRMCosmosdbCassandraKeyspaceRegistryItem(),
	GetAzureRMCosmosdbCassandraTableRegistryItem(),
	GetAzureRMCosmosdbGremlinDatabaseRegistryItem(),
	GetAzureRMCosmosdbGremlinGraphRegistryItem(),
	GetAzureRMCosmosdbMongoCollectionRegistryItem(),
	GetAzureRMCosmosdbMongoDatabaseRegistryItem(),
	GetAzureRMCosmosdbSQLContainerRegistryItem(),
	GetAzureRMCosmosdbSQLDatabaseRegistryItem(),
	GetAzureRMCosmosdbTableRegistryItem(),
	getDatabricksWorkspaceRegistryItem(),
	getDNSARecordRegistryItem(),
	getDNSAAAARecordRegistryItem(),
	getDNSCAARecordRegistryItem(),
	getDNSCNameRecordRegistryItem(),
	getDNSMXRecordRegistryItem(),
	getDNSNSRecordRegistryItem(),
	getDNSPtrRecordRegistryItem(),
	getDNSSrvRecordRegistryItem(),
	getDNSTxtRecordRegistryItem(),
	getDNSPrivateZoneRegistryItem(),
	getDNSZoneRegistryItem(),
	GetAzureRMEventHubsNamespaceRegistryItem(),
	getExpressRouteConnectionRegistryItem(),
	getExpressRouteGatewayRegistryItem(),
	GetAzureRMFirewallRegistryItem(),
	getFrontdoorFirewallPolicyRegistryItem(),
	getFrontdoorRegistryItem(),
	GetAzureRMHDInsightHadoopClusterRegistryItem(),
	GetAzureRMHDInsightHBaseClusterRegistryItem(),
	GetAzureRMHDInsightInteractiveQueryClusterRegistryItem(),
	GetAzureRMHDInsightKafkaClusterRegistryItem(),
	GetAzureRMHDInsightSparkClusterRegistryItem(),
	GetAzureRMKeyVaultCertificateRegistryItem(),
	GetAzureRMKeyVaultKeyRegistryItem(),
	GetAzureRMKeyVaultManagedHSMRegistryItem(),
	getKubernetesClusterRegistryItem(),
	getKubernetesClusterNodePoolRegistryItem(),
	getLoadBalancerRegistryItem(),
	GetAzureRMLoadBalancerRuleRegistryItem(),
	GetAzureRMLoadBalancerOutboundRuleRegistryItem(),
	getLinuxFunctionAppRegistryItem(),
	getLinuxVirtualMachineRegistryItem(),
	getLinuxVirtualMachineScaleSetRegistryItem(),
	getLogAnalyticsWorkspaceRegistryItem(),
	getManagedDiskRegistryItem(),
	GetAzureRMMariaDBServerRegistryItem(),
	getMSSQLDatabaseRegistryItem(),
	GetAzureRMMySQLServerRegistryItem(),
	GetAzureRMNotificationHubNamespaceRegistryItem(),
	getPointToSiteVpnGatewayRegistryItem(),
	getPostgreSQLFlexibleServerRegistryItem(),
	GetAzureRMPostgreSQLServerRegistryItem(),
	getPrivateDNSARecordRegistryItem(),
	getPrivateDNSAAAARecordRegistryItem(),
	getPrivateDNSCNameRecordRegistryItem(),
	getPrivateDNSMXRecordRegistryItem(),
	getPrivateDNSPTRRecordRegistryItem(),
	getPrivateDNSSRVRecordRegistryItem(),
	getPrivateDNSTXTRecordRegistryItem(),
	GetAzureRMPrivateEndpointRegistryItem(),
	GetAzureRMPublicIPRegistryItem(),
	GetAzureRMPublicIPPrefixRegistryItem(),
	GetAzureRMSearchServiceRegistryItem(),
	GetAzureRMRedisCacheRegistryItem(),
	getAzureRMMSSQLManagedInstanceRegistryItem(),
	getStorageAccountRegistryItem(),
	getSQLDatabaseRegistryItem(),
	getSQLManagedInstanceRegistryItem(),
	GetAzureRMSynapseSparkPoolRegistryItem(),
	GetAzureRMSynapseSQLPoolRegistryItem(),
	GetAzureRMSynapseWorkspacRegistryItem(),
	getVirtualHubRegistryItem(),
	getVirtualMachineScaleSetRegistryItem(),
	getVirtualMachineRegistryItem(),
	GetAzureRMVirtualNetworkGatewayConnectionRegistryItem(),
	GetAzureRMVirtualNetworkGatewayRegistryItem(),
	getWindowsVirtualMachineRegistryItem(),
	getWindowsVirtualMachineScaleSetRegistryItem(),
	getVPNGatewayRegistryItem(),
	getVPNGatewayConnectionRegistryItem(),
	getDataFactoryRegistryItem(),
	getDataFactoryIntegrationRuntimeAzureRegistryItem(),
	getDataFactoryIntegrationRuntimeAzureSSISRegistryItem(),
	getDataFactoryIntegrationRuntimeManagedRegistryItem(),
	getDataFactoryIntegrationRuntimeSelfHostedRegistryItem(),
	getLogAnalyticsSolutionRegistryItem(),
	getMySQLFlexibleServerRegistryItem(),
	getServicePlanRegistryItem(),
	getSentinelDataConnectorAwsCloudTrailRegistryItem(),
	getSentinelDataConnectorAzureActiveDirectoryRegistryItem(),
	getSentinelDataConnectorAzureAdvancedThreatProtectionRegistryItem(),
	getSentinelDataConnectorAzureSecurityCenterRegistryItem(),
	getSentinelDataConnectorMicrosoftCloudAppSecurityRegistryItem(),
	getSentinelDataConnectorMicrosoftDefenderAdvancedThreatProtectionRegistryItem(),
	getSentinelDataConnectorOffice365RegistryItem(),
	getSentinelDataConnectorThreatIntelligenceRegistryItem(),
	getIoTHubRegistryItem(),
	getIoTHubDPSRegistryItem(),
	getVirtualNetworkPeeringRegistryItem(),
	geWindowsFunctionAppRegistryItem(),
	getPowerBIEmbeddedRegistryItem(),
	getMSSQLElasticPoolRegistryItem(),
	getSQLElasticPoolRegistryItem(),
	getMonitorActionGroupRegistryItem(),
	getMonitorDataCollectionRuleRegistryItem(),
	getMonitorDiagnosticSettingRegistryItem(),
	getMonitorMetricAlertRegistryItem(),
	getMonitorScheduledQueryRulesAlertRegistryItem(),
	getMonitorScheduledQueryRulesAlertV2RegistryItem(),
	getApplicationInsightsStandardWebTestRegistryItem(),
	getRecoveryServicesVaultRegistryItem(),
	getBackupProtectedVmRegistryItem(),
	getStorageManagementPolicyRegistryItem(),
	getStorageQueueRegistryItem(),
	getStorageShareRegistryItem(),
	getLogicAppIntegrationAccountRegistryItem(),
	getSignalRServiceRegistryItem(),
	getTrafficManagerProfileRegistryItem(),
	getTrafficManagerAzureEndpointRegistryItem(),
	getTrafficManagerExternalEndpointRegistryItem(),
	getTrafficManagerNestedEndpointRegistryItem(),
	getEventgridSystemTopicRegistryItem(),
	getEventgridTopicRegistryItem(),
	getSecurityCenterSubscriptionPricingRegistryItem(),
	getNetworkWatcherFlowLogRegistryItem(),
	getNetworkWatcherRegistryItem(),
	getNetworkConnectionMonitorRegistryItem(),
	getServiceBusNamespaceRegistryItem(),
	getLogicAppStandardRegistryItem(),
	getImageRegistryItem(),
	getSnapshotRegistryItem(),
	getPrivateDnsResolverInboundEndpointRegistryItem(),
	getPrivateDnsResolverOutboundEndpointRegistryItem(),
	getPrivateDnsResolverDnsForwardingRulesetRegistryItem(),
	getMachineLearningComputeInstanceRegistryItem(),
	getMachineLearningComputeClusterRegistryItem(),
	getNetworkDdosProtectionPlanRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources = []string{
	// Azure Api Management
	"azurerm_api_management_api",
	"azurerm_api_management_api_diagnostic",
	"azurerm_api_management_api_operation",
	"azurerm_api_management_api_operation_policy",
	"azurerm_api_management_api_policy",
	"azurerm_api_management_api_schema",
	"azurerm_api_management_api_version_set",
	"azurerm_api_management_authorization_server",
	"azurerm_api_management_backend",
	"azurerm_api_management_certificate",
	"azurerm_api_management_custom_domain",
	"azurerm_api_management_diagnostic",
	"azurerm_api_management_email_template",
	"azurerm_api_management_group",
	"azurerm_api_management_group_user",
	"azurerm_api_management_identity_provider_aad",
	"azurerm_api_management_identity_provider_aadb2c",
	"azurerm_api_management_identity_provider_facebook",
	"azurerm_api_management_identity_provider_google",
	"azurerm_api_management_identity_provider_microsoft",
	"azurerm_api_management_identity_provider_twitter",
	"azurerm_api_management_logger",
	"azurerm_api_management_named_value",
	"azurerm_api_management_openid_connect_provider",
	"azurerm_api_management_policy",
	"azurerm_api_management_product",
	"azurerm_api_management_product_api",
	"azurerm_api_management_product_group",
	"azurerm_api_management_product_policy",
	"azurerm_api_management_property",
	"azurerm_api_management_subscription",
	"azurerm_api_management_user",

	// Azure Application Gateway
	"azurerm_web_application_firewall_policy",

	// Azure App Service
	"azurerm_app_service",
	"azurerm_app_service_active_slot",
	"azurerm_app_service_certificate",
	"azurerm_app_service_managed_certificate",
	"azurerm_app_service_slot",
	"azurerm_app_service_slot_virtual_network_swift_connection",
	"azurerm_app_service_source_control_token",
	"azurerm_app_service_virtual_network_swift_connection",

	// Azure Attestation
	"azurerm_attestation_provider",

	// Azure Automation
	"azurerm_automation_certificate",
	"azurerm_automation_connection",
	"azurerm_automation_connection_certificate",
	"azurerm_automation_connection_type",
	"azurerm_automation_connection_classic_certificate",
	"azurerm_automation_connection_service_principal",
	"azurerm_automation_credential",
	"azurerm_automation_hybrid_runbook_worker",
	"azurerm_automation_hybrid_runbook_worker_group",
	"azurerm_automation_module",
	"azurerm_automation_runbook",
	"azurerm_automation_schedule",
	"azurerm_automation_software_update_configuration",
	"azurerm_automation_source_control",
	"azurerm_automation_variable_bool",
	"azurerm_automation_variable_datetime",
	"azurerm_automation_variable_int",
	"azurerm_automation_variable_string",
	"azurerm_automation_webhook",

	// Azure Backup & Recovery Services Vault
	"azurerm_backup_policy_vm",
	"azurerm_backup_policy_file_share",
	"azurerm_site_recovery_network_mapping",
	"azurerm_site_recovery_replication_policy",

	// Azure Base
	"azurerm_resource_group",
	"azurerm_resource_provider_registration",
	"azurerm_subscription",
	"azurerm_role_assignment",
	"azurerm_role_definition",
	"azurerm_user_assigned_identity",

	// Azure Blueprints
	"azurerm_blueprint_assignment",

	// Azure CDN
	"azurerm_cdn_profile",

	// Azure Consumption
	"azurerm_consumption_budget_management_group",
	"azurerm_consumption_budget_resource_group",
	"azurerm_consumption_budget_subscription",

	// Azure CosmosDB
	"azurerm_cosmosdb_notebook_workspace",
	"azurerm_cosmosdb_sql_stored_procedure",
	"azurerm_cosmosdb_sql_trigger",
	"azurerm_cosmosdb_sql_user_defined_function",

	// Azure Cost Management
	"azurerm_cost_anomaly_alert",
	"azurerm_cost_management_scheduled_action",
	"azurerm_resource_group_cost_management_export",
	"azurerm_resource_group_cost_management_view",
	"azurerm_subscription_cost_management_export",
	"azurerm_subscription_cost_management_view",

	// Azure DNS
	"azurerm_private_dns_zone_virtual_network_link",
	"azurerm_private_dns_resolver",

	// Azure Data Factory
	"azurerm_data_factory_custom_dataset",
	"azurerm_data_factory_data_flow",
	"azurerm_data_factory_dataset_azure_blob",
	"azurerm_data_factory_dataset_binary",
	"azurerm_data_factory_dataset_cosmosdb_sqlapi",
	"azurerm_data_factory_dataset_delimited_text",
	"azurerm_data_factory_dataset_http",
	"azurerm_data_factory_dataset_json",
	"azurerm_data_factory_dataset_mysql",
	"azurerm_data_factory_dataset_parquet",
	"azurerm_data_factory_dataset_postgresql",
	"azurerm_data_factory_dataset_snowflake",
	"azurerm_data_factory_dataset_sql_server_table",
	"azurerm_data_factory_linked_custom_service",
	"azurerm_data_factory_linked_service_azure_blob_storage",
	"azurerm_data_factory_linked_service_azure_databricks",
	"azurerm_data_factory_linked_service_azure_file_storage",
	"azurerm_data_factory_linked_service_azure_function",
	"azurerm_data_factory_linked_service_azure_search",
	"azurerm_data_factory_linked_service_azure_sql_database",
	"azurerm_data_factory_linked_service_azure_table_storage",
	"azurerm_data_factory_linked_service_cosmosdb",
	"azurerm_data_factory_linked_service_cosmosdb_mongoapi",
	"azurerm_data_factory_linked_service_data_lake_storage_gen2",
	"azurerm_data_factory_linked_service_key_vault",
	"azurerm_data_factory_linked_service_kusto",
	"azurerm_data_factory_linked_service_mysql",
	"azurerm_data_factory_linked_service_odata",
	"azurerm_data_factory_linked_service_odbc",
	"azurerm_data_factory_linked_service_postgresql",
	"azurerm_data_factory_linked_service_sftp",
	"azurerm_data_factory_linked_service_snowflake",
	"azurerm_data_factory_linked_service_sql_server",
	"azurerm_data_factory_linked_service_synapse",
	"azurerm_data_factory_linked_service_web",
	"azurerm_data_factory_managed_private_endpoint",
	"azurerm_data_factory_pipeline",
	"azurerm_data_factory_trigger_blob_event",
	"azurerm_data_factory_trigger_custom_event",
	"azurerm_data_factory_trigger_schedule",
	"azurerm_data_factory_tumbling_window",

	// Azure Database
	"azurerm_mariadb_configuration",
	"azurerm_mariadb_database",
	"azurerm_mariadb_firewall_rule",
	"azurerm_mariadb_virtual_network_rule",

	"azurerm_mysql_active_directory_administrator",
	"azurerm_mysql_configuration",
	"azurerm_mysql_database",
	"azurerm_mysql_firewall_rule",
	"azurerm_mysql_flexible_database",
	"azurerm_mysql_flexible_server_configuration",
	"azurerm_mysql_flexible_server_firewall_rule",
	"azurerm_mysql_server_key",
	"azurerm_mysql_virtual_network_rule",

	"azurerm_postgresql_active_directory_administrator",
	"azurerm_postgresql_configuration",
	"azurerm_postgresql_database",
	"azurerm_postgresql_firewall_rule",
	"azurerm_postgresql_flexible_server_active_directory_administrator",
	"azurerm_postgresql_flexible_server_configuration",
	"azurerm_postgresql_flexible_server_database",
	"azurerm_postgresql_flexible_server_firewall_rule",
	"azurerm_postgresql_server_key",
	"azurerm_postgresql_virtual_network_rule",

	// Azure Datalake Gen 2
	"azurerm_storage_data_lake_gen2_filesystem",

	// Azure Event Grid
	"azurerm_eventgrid_domain",
	"azurerm_eventgrid_event_subscription",
	"azurerm_eventgrid_system_topic_event_subscription",

	// Azure Event Hub
	"azurerm_eventhub",
	"azurerm_eventhub_authorization_rule",
	"azurerm_eventhub_cluster",
	"azurerm_eventhub_consumer_group",
	"azurerm_eventhub_namespace_authorization_rule",
	"azurerm_eventhub_namespace_customer_managed_key",
	"azurerm_eventhub_namespace_disaster_recovery_config",

	// Azure Firewall
	"azurerm_firewall_application_rule_collection",
	"azurerm_firewall_nat_rule_collection",
	"azurerm_firewall_network_rule_collection",
	"azurerm_firewall_policy",
	"azurerm_firewall_policy_rule_collection_group",

	// Azure Front Door
	"azurerm_frontdoor_custom_https_configuration",
	"azurerm_frontdoor_rules_engine",

	// Azure Key Vault
	"azurerm_key_vault",
	"azurerm_key_vault_access_policy",
	"azurerm_key_vault_certificate_data",
	"azurerm_key_vault_certificate_issuer",
	"azurerm_key_vault_secret",

	// Azure IoT
	"azurerm_iothub_certificate",
	"azurerm_iothub_consumer_group",
	"azurerm_iothub_dps_certificate",
	"azurerm_iothub_dps_shared_access_policy",
	"azurerm_iothub_endpoint_eventhub",
	"azurerm_iothub_enrichment",
	"azurerm_iothub_route",
	"azurerm_iothub_shared_access_policy",

	// Azure Lighthouse (Delegated Resoure Management)
	"azurerm_lighthouse_definition",
	"azurerm_lighthouse_assignment",

	// Azure Load Balancer
	"azurerm_lb_backend_address_pool",
	"azurerm_lb_backend_address_pool_address",
	"azurerm_lb_nat_pool",
	"azurerm_lb_nat_rule",
	"azurerm_lb_probe",

	// Azure Logic App
	"azurerm_logic_app_action_custom",
	"azurerm_logic_app_action_http",
	"azurerm_logic_app_integration_account_agreement",
	"azurerm_logic_app_integration_account_assembly",
	"azurerm_logic_app_integration_account_batch_configuration",
	"azurerm_logic_app_integration_account_certificate",
	"azurerm_logic_app_integration_account_map",
	"azurerm_logic_app_integration_account_partner",
	"azurerm_logic_app_integration_account_schema",
	"azurerm_logic_app_integration_account_session",
	"azurerm_logic_app_trigger_custom",
	"azurerm_logic_app_trigger_http_request",
	"azurerm_logic_app_trigger_recurrence",
	"azurerm_logic_app_workflow",

	// Azure Machine Learning
	"azurerm_machine_learning_workspace",

	// Azure Management
	"azurerm_management_group",
	"azurerm_management_group_subscription_association",
	"azurerm_management_group_policy_assignment",
	"azurerm_management_lock",

	// Azure Managed Applications
	"azurerm_managed_application",
	"azurerm_managed_application_definition",

	// Azure Monitor
	"azurerm_monitor_aad_diagnostic_setting",
	"azurerm_monitor_action_rule_action_group",
	"azurerm_monitor_action_rule_suppression",
	"azurerm_monitor_activity_log_alert",
	"azurerm_monitor_alert_processing_rule_action_group",
	"azurerm_monitor_alert_processing_rule_suppression",
	"azurerm_monitor_autoscale_setting",
	"azurerm_monitor_data_collection",
	"azurerm_monitor_data_collection_rule_association",
	"azurerm_monitor_log_profile",
	"azurerm_monitor_private_link_scope",
	"azurerm_monitor_private_link_scoped_service",
	"azurerm_monitor_scheduled_query_rules_log",
	"azurerm_monitor_smart_detector_alert_rule",

	// Azure Monitor - Application Insights
	"azurerm_application_insights_analytics_item",
	"azurerm_application_insights_api_key",
	"azurerm_application_insights_smart_detection_rule",
	"azurerm_application_insights_workbook",
	"azurerm_application_insights_workbook_template",

	// Azure Monitor - Log Analytics
	"azurerm_log_analytics_cluster_customer_managed_key",
	"azurerm_log_analytics_data_export_rule",
	"azurerm_log_analytics_datasource_windows_event",
	"azurerm_log_analytics_datasource_windows_performance_counter",
	"azurerm_log_analytics_linked_service",
	"azurerm_log_analytics_linked_storage_account",
	"azurerm_log_analytics_query_pack",
	"azurerm_log_analytics_query_pack_query",
	"azurerm_log_analytics_saved_search",
	"azurerm_log_analytics_storage_insights",

	// Azure Networking
	"azurerm_application_security_group",
	"azurerm_ip_group",
	"azurerm_local_network_gateway",
	"azurerm_nat_gateway_public_ip_association",
	"azurerm_nat_gateway_public_ip_prefix_association",
	"azurerm_network_interface",
	"azurerm_network_interface_application_gateway_backend_address_pool_association",
	"azurerm_network_interface_application_security_group_association",
	"azurerm_network_interface_backend_address_pool_association",
	"azurerm_network_interface_nat_rule_association",
	"azurerm_network_interface_security_group_association",
	"azurerm_network_security_group",
	"azurerm_network_security_rule",
	"azurerm_private_link_service",
	"azurerm_route",
	"azurerm_route_filter",
	"azurerm_route_map",
	"azurerm_route_table",
	"azurerm_storage_account_local_user",
	"azurerm_storage_account_network_rules",
	"azurerm_subnet",
	"azurerm_subnet_nat_gateway_association",
	"azurerm_subnet_network_security_group_association",
	"azurerm_subnet_route_table_association",
	"azurerm_subnet_service_endpoint_storage_policy",
	"azurerm_virtual_network",
	"azurerm_virtual_network_dns_servers",

	// Azure Notification Hub
	"azurerm_notification_hub",

	// Azure Policy
	"azurerm_policy_assignment",
	"azurerm_policy_definition",
	"azurerm_policy_remediation",
	"azurerm_policy_set_definition",
	"azurerm_subscription_policy_assignment",
	"azurerm_subscription_policy_exemption",
	"azurerm_subscription_policy_remediation",

	// Azure Portal
	"azurerm_dashboard",
	"azurerm_portal_dashboard",

	// Azure Redis
	"azurerm_redis_firewall_rule",
	"azurerm_redis_linked_server",

	// Azure Registry
	"azurerm_container_registry_scope_map",
	"azurerm_container_registry_token",
	"azurerm_container_registry_webhook",

	// Azure Sentinel
	"azurerm_sentinel_alert_rule_machine_learning_behavior_analytics",
	"azurerm_sentinel_alert_rule_fusion",
	"azurerm_sentinel_alert_rule_ms_security_incident",
	"azurerm_sentinel_alert_rule_scheduled",

	// Azure Service Bus
	"azurerm_servicebus_namespace_authorization_rule",
	"azurerm_servicebus_namespace_disaster_recovery_config",
	"azurerm_servicebus_namespace_network_rule_set",
	"azurerm_servicebus_queue",
	"azurerm_servicebus_queue_authorization_rule",
	"azurerm_servicebus_subscription",
	"azurerm_servicebus_subscription_rule",
	"azurerm_servicebus_topic",
	"azurerm_servicebus_topic_authorization_rule",
	"azurerm_relay_hybrid_connection_authorization_rule",
	"azurerm_relay_namespace_authorization_rule",

	// Azure Shared Image Gallery
	"azurerm_shared_image",
	"azurerm_shared_image_gallery",

	// Azure SignalR
	"azurerm_signalr_service_network_acl",
	"azurerm_signalr_shared_private_link",

	// Azure SQL
	"azurerm_sql_failover_group",
	"azurerm_sql_firewall_rule",
	"azurerm_sql_server",
	"azurerm_sql_virtual_network_rule",

	"azurerm_mssql_database_extended_auditing_policy",
	"azurerm_mssql_database_vulnerability_assessment_rule_baseline",
	"azurerm_mssql_failover_group",
	"azurerm_mssql_firewall_rule",
	"azurerm_mssql_job_agent",
	"azurerm_mssql_job_credential",
	"azurerm_mssql_managed_instance_active_directory_administrator",
	"azurerm_mssql_managed_instance_security_alert_policy",
	"azurerm_mssql_managed_instance_transparent_data_encryption",
	"azurerm_mssql_managed_instance_vulnerability_assessment",
	"azurerm_mssql_outbound_firewall_rule",
	"azurerm_mssql_server",
	"azurerm_mssql_server_dns_alias",
	"azurerm_mssql_server_extended_auditing_policy",
	"azurerm_mssql_server_microsoft_support_auditing_policy",
	"azurerm_mssql_server_security_alert_policy",
	"azurerm_mssql_server_transparent_data_encryption",
	"azurerm_mssql_server_vulnerability_assessment",
	"azurerm_mssql_virtual_network_rule",

	// Azure Storage
	"azurerm_storage_account_customer_managed_key",
	"azurerm_storage_account_local_user",
	"azurerm_storage_account_network_rules",
	"azurerm_storage_blob",
	"azurerm_storage_blob_inventory_policy",
	"azurerm_storage_container",
	"azurerm_storage_data_lake_gen2_path",
	"azurerm_storage_object_replication",
	"azurerm_storage_share_directory",
	"azurerm_storage_share_file",
	"azurerm_storage_sync_cloud_endpoint",
	"azurerm_storage_sync_group",
	"azurerm_storage_table_entity",

	// Azure Virtual Desktop
	"azurerm_virtual_desktop_application",
	"azurerm_virtual_desktop_application_group",
	"azurerm_virtual_desktop_workspace",
	"azurerm_virtual_desktop_workspace_application_group_association",
	"azurerm_virtual_desktop_host_pool",

	// Azure Service Plan
	"azurerm_windows_web_app",
	"azurerm_linux_web_app",

	// Azure Synapse Analytics
	"azurerm_synapse_firewall_rule",
	"azurerm_synapse_private_link_hub",

	// Azure Virtual Hub
	"azurerm_virtual_hub_route_table",
	"azurerm_virtual_hub_route_table_route",

	// Azure Virtual Machines
	"azurerm_virtual_machine_data_disk_attachment",
	"azurerm_virtual_machine_extension",
	"azurerm_virtual_machine_scale_set_extension",
	"azurerm_availability_set",
	"azurerm_proximity_placement_group",
	"azurerm_ssh_public_key",
	"azurerm_marketplace_agreement",

	// Azure WAN
	"azurerm_virtual_hub_connection",
	"azurerm_virtual_wan",
	"azurerm_vpn_server_configuration",

	// Microsoft Defender for Cloud
	"azurerm_security_center_automation",
	"azurerm_security_center_server_vulnerability_assessment",
	"azurerm_security_center_assessment",
	"azurerm_security_center_assessment_policy",
	"azurerm_security_center_auto_provisioning",
	"azurerm_security_center_automation",
	"azurerm_security_center_contact",
	"azurerm_security_center_server_vulnerability_assessment_virtual_machine",
	"azurerm_security_center_setting",
	"azurerm_security_center_workspace",
}

var UsageOnlyResources = []string{}
