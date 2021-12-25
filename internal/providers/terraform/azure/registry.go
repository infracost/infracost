package azure

import "github.com/infracost/infracost/internal/schema"

// ResourceRegistry grouped alphabetically
var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMActiveDirectoryDomainServiceRegistryItem(),
	GetAzureRMActiveDirectoryDomainServiceReplicaSetRegistryItem(),
	GetAzureRMApiManagementRegistryItem(),
	GetAzureRMApplicationGatewayRegistryItem(),
	GetAzureRMAppIsolatedServicePlanRegistryItem(),
	GetAzureRMAppIntegrationServiceEnvironmentRegistryItem(),
	GetAzureRMAppFunctionRegistryItem(),
	GetAzureRMAppNATGatewayRegistryItem(),
	GetAzureRMAppServiceCertificateBindingRegistryItem(),
	GetAzureRMAppServiceCertificateOrderRegistryItem(),
	GetAzureRMAppServiceCustomHostnameBindingRegistryItem(),
	GetAzureRMAppServicePlanRegistryItem(),
	GetAzureRMApplicationInsightsWebRegistryItem(),
	GetAzureRMApplicationInsightsRegistryItem(),
	GetAzureRMAutomationAccountRegistryItem(),
	GetAzureRMAutomationDscConfigurationRegistryItem(),
	GetAzureRMAutomationDscNodeconfigurationRegistryItem(),
	GetAzureRMAutomationJobScheduleRegistryItem(),
	GetAzureRMBastionHostRegistryItem(),
	GetAzureRMCDNEndpointRegistryItem(),
	GetAzureRMContainerRegistryRegistryItem(),
	GetAzureRMCosmosdbCassandraKeyspaceRegistryItem(),
	GetAzureRMCosmosdbCassandraTableRegistryItem(),
	GetAzureRMCosmosdbGremlinDatabaseRegistryItem(),
	GetAzureRMCosmosdbGremlinGraphRegistryItem(),
	GetAzureRMCosmosdbMongoCollectionRegistryItem(),
	GetAzureRMCosmosdbMongoDatabaseRegistryItem(),
	GetAzureRMCosmosdbSQLContainerRegistryItem(),
	GetAzureRMCosmosdbSQLDatabaseRegistryItem(),
	GetAzureRMCosmosdbTableRegistryItem(),
	GetAzureRMDatabricksWorkspaceRegistryItem(),
	GetAzureRMDNSaRecordRegistryItem(),
	GetAzureRMDNSaaaaRecordRegistryItem(),
	GetAzureRMDNScaaRecordRegistryItem(),
	GetAzureRMDNScnameRecordRegistryItem(),
	GetAzureRMDNSmxRecordRegistryItem(),
	GetAzureRMDNSnsRecordRegistryItem(),
	GetAzureRMDNSptrRecordRegistryItem(),
	GetAzureRMDNSsrvRecordRegistryItem(),
	GetAzureRMDNStxtRecordRegistryItem(),
	GetAzureRMDNSPrivateZoneRegistryItem(),
	GetAzureRMDNSZoneRegistryItem(),
	GetAzureRMEventHubsNamespaceRegistryItem(),
	getAzureRMExpressRouteConnectionRegistryItem(),
	getAzureRMExpressRouteGatewayRegistryItem(),
	GetAzureRMFirewallRegistryItem(),
	getAzureRMFrontdoorFirewallPolicyRegistryItem(),
	getAzureRMFrontdoorRegistryItem(),
	GetAzureRMHDInsightHadoopClusterRegistryItem(),
	GetAzureRMHDInsightHBaseClusterRegistryItem(),
	GetAzureRMHDInsightInteractiveQueryClusterRegistryItem(),
	GetAzureRMHDInsightKafkaClusterRegistryItem(),
	GetAzureRMHDInsightSparkClusterRegistryItem(),
	GetAzureRMKeyVaultCertificateRegistryItem(),
	GetAzureRMKeyVaultKeyRegistryItem(),
	GetAzureRMKeyVaultManagedHSMRegistryItem(),
	GetAzureRMKubernetesClusterRegistryItem(),
	GetAzureRMKubernetesClusterNodePoolRegistryItem(),
	GetAzureRMLoadBalancerRegistryItem(),
	GetAzureRMLoadBalancerRuleRegistryItem(),
	GetAzureRMLoadBalancerOutboundRuleRegistryItem(),
	GetAzureRMLinuxVirtualMachineRegistryItem(),
	GetAzureRMLinuxVirtualMachineScaleSetRegistryItem(),
	GetAzureRMManagedDiskRegistryItem(),
	GetAzureRMMariaDBServerRegistryItem(),
	getAzureRMMSSQLDatabaseRegistryItem(),
	GetAzureRMMySQLServerRegistryItem(),
	GetAzureRMNotificationHubNamespaceRegistryItem(),
	getAzureRMPointToSiteVpnGatewayRegistryItem(),
	GetAzureRMPostgreSQLFlexibleServerRegistryItem(),
	GetAzureRMPostgreSQLServerRegistryItem(),
	GetAzureRMPrivateDNSaRecordRegistryItem(),
	GetAzureRMPrivateDNSaaaaRecordRegistryItem(),
	GetAzureRMPrivateDNScnameRecordRegistryItem(),
	GetAzureRMPrivateDNSmxRecordRegistryItem(),
	GetAzureRMPrivateDNSptrRecordRegistryItem(),
	GetAzureRMPrivateDNSsrvRecordRegistryItem(),
	GetAzureRMPrivateDNStxtRecordRegistryItem(),
	GetAzureRMPrivateEndpointRegistryItem(),
	GetAzureRMPublicIPRegistryItem(),
	GetAzureRMPublicIPPrefixRegistryItem(),
	GetAzureRMSearchServiceRegistryItem(),
	GetAzureRMRedisCacheRegistryItem(),
	getAzureRMStorageAccountRegistryItem(),
	getAzureRMSQLDatabaseRegistryItem(),
	GetAzureRMSynapseSparkPoolRegistryItem(),
	GetAzureRMSynapseSQLPoolRegistryItem(),
	GetAzureRMSynapseWorkspacRegistryItem(),
	getAzureRMVirtualHubRegistryItem(),
	GetAzureRMVirtualMachineScaleSetRegistryItem(),
	GetAzureRMVirtualMachineRegistryItem(),
	GetAzureRMVirtualNetworkGatewayConnectionRegistryItem(),
	GetAzureRMVirtualNetworkGatewayRegistryItem(),
	GetAzureRMWindowsVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineScaleSetRegistryItem(),
	getAzureRMVPNGatewayRegistryItem(),
	getAzureRMVPNGatewayConnectionRegistryItem(),
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

	// Azure App Service
	"azurerm_app_service_active_slot",
	"azurerm_app_service_certificate",
	"azurerm_app_service_managed_certificate",
	"azurerm_app_service_slot",
	"azurerm_app_service_slot_virtual_network_swift_connection",
	"azurerm_app_service_source_control_token",
	"azurerm_app_service_virtual_network_swift_connection",

	// Azure Automation
	"azurerm_automation_certificate",
	"azurerm_automation_connection",
	"azurerm_automation_connection_certificate",
	"azurerm_automation_connection_classic_certificate",
	"azurerm_automation_connection_service_principal",
	"azurerm_automation_credential",
	"azurerm_automation_module",
	"azurerm_automation_runbook",
	"azurerm_automation_schedule",
	"azurerm_automation_variable_bool",
	"azurerm_automation_variable_datetime",
	"azurerm_automation_variable_int",
	"azurerm_automation_variable_string",

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

	// Azure CosmosDB
	"azurerm_cosmosdb_account",
	"azurerm_cosmosdb_notebook_workspace",
	"azurerm_cosmosdb_sql_stored_procedure",
	"azurerm_cosmosdb_sql_trigger",
	"azurerm_cosmosdb_sql_user_defined_function",

	// Azure DNS
	"azurerm_private_dns_zone_virtual_network_link",

	// Azure Database
	"azurerm_mariadb_configuration",
	"azurerm_mariadb_firewall_rule",
	"azurerm_mariadb_virtual_network_rule",
	"azurerm_mysql_firewall_rule",
	"azurerm_mysql_virtual_network_rule",
	"azurerm_postgresql_configuration",
	"azurerm_postgresql_firewall_rule",
	"azurerm_postgresql_flexible_server_firewall_rule",
	"azurerm_postgresql_virtual_network_rule",

	// Azure Datalake Gen 2
	"azurerm_storage_data_lake_gen2_filesystem",

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

	// Azure Lighthouse (Delegated Resoure Management)
	"azurerm_lighthouse_definition",
	"azurerm_lighthouse_assignment",

	// Azure Load Balancer
	"azurerm_lb_backend_address_pool",
	"azurerm_lb_backend_address_pool_address",
	"azurerm_lb_nat_pool",
	"azurerm_lb_nat_rule",
	"azurerm_lb_probe",

	// Azure Management
	"azurerm_management_group",
	"azurerm_management_group_subscription_association",
	"azurerm_management_group_policy_assignment",
	"azurerm_management_lock",

	// Azure Managed Applications
	"azurerm_managed_application",
	"azurerm_managed_application_definition",

	// Azure Networking
	"azurerm_application_security_group",
	"azurerm_local_network_gateway",
	"azurerm_nat_gateway_public_ip_association",
	"azurerm_nat_gateway_public_ip_prefix_association",
	"azurerm_network_interface",
	"azurerm_network_interface_security_group_association",
	"azurerm_network_security_group",
	"azurerm_network_security_rule",
	"azurerm_private_link_service",
	"azurerm_storage_account_network_rules",
	"azurerm_subnet",
	"azurerm_subnet_network_security_group_association",
	"azurerm_virtual_network",

	// Azure Notification Hub
	"azurerm_notification_hub",

	// Azure Policy
	"azurerm_policy_assignment",
	"azurerm_policy_definition",
	"azurerm_policy_remediation",
	"azurerm_policy_set_definition",

	// Azure Portal
	"azurerm_dashboard",

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

	// Azure SQL
	"azurerm_sql_server",
	"azurerm_sql_firewall_rule",
	"azurerm_sql_virtual_network_rule",
	"azurerm_mssql_firewall_rule",

	// Azure Storage
	"azurerm_storage_table_entity",
	"azurerm_storage_management_policy",
	"azurerm_storage_blob_inventory_policy",

	// Azure Virtual Desktop
	"azurerm_virtual_desktop_application",
	"azurerm_virtual_desktop_application_group",
	"azurerm_virtual_desktop_workspace",
	"azurerm_virtual_desktop_workspace_application_group_association",
	"azurerm_virtual_desktop_host_pool",

	// Azure Synapse Analytics
	"azurerm_synapse_firewall_rule",
	"azurerm_synapse_private_link_hub",

	// Azure Virtual Machines
	"azurerm_virtual_machine_data_disk_attachment",
	"azurerm_availability_set",
	"azurerm_proximity_placement_group",
	"azurerm_ssh_public_key",
	"azurerm_marketplace_agreement",

	// Azure WAN
	"azurerm_virtual_wan",
}

var UsageOnlyResources = []string{}
