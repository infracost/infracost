package azure

import "github.com/infracost/infracost/internal/schema"

// ResourceRegistry grouped alphabetically
var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
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
	GetAzureRMFirewallRegistryItem(),
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
	GetAzureRMMSSQLDatabaseRegistryItem(),
	GetAzureRMMySQLServerRegistryItem(),
	GetAzureRMNotificationHubNamespaceRegistryItem(),
	GetAzureRMPostgreSQLFlexibleServerRegistryItem(),
	GetAzureRMPostgreSQLServerRegistryItem(),
	GetAzureRMPrivateDNSaRecordRegistryItem(),
	GetAzureRMPrivateDNSaaaaRecordRegistryItem(),
	GetAzureRMPrivateDNScnameRecordRegistryItem(),
	GetAzureRMPrivateDNSmxRecordRegistryItem(),
	GetAzureRMPrivateDNSptrRecordRegistryItem(),
	GetAzureRMPrivateDNSsrvRecordRegistryItem(),
	GetAzureRMPrivateDNStxtRecordRegistryItem(),
	GetAzureRMPublicIPRegistryItem(),
	GetAzureRMPublicIPPrefixRegistryItem(),
	GetAzureRMSearchServiceRegistryItem(),
	GetAzureRMRedisCacheRegistryItem(),
	GetAzureRMStorageAccountRegistryItem(),
	GetAzureRMVirtualMachineScaleSetRegistryItem(),
	GetAzureRMVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineScaleSetRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
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

	// Azure Base
	"azurerm_resource_group",
	"azurerm_resource_provider_registration",
	"azurerm_subscription",
	"azurerm_role_assignment",

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

	// Azure Key Vault
	"azurerm_key_vault",
	"azurerm_key_vault_access_policy",
	"azurerm_key_vault_certificate_data",
	"azurerm_key_vault_certificate_issuer",
	"azurerm_key_vault_secret",

	// Azure Load Balancer
	"azurerm_lb_backend_address_pool",
	"azurerm_lb_backend_address_pool_address",
	"azurerm_lb_nat_pool",
	"azurerm_lb_nat_rule",
	"azurerm_lb_probe",

	// Azure Networking
	"azurerm_application_security_group",
	"azurerm_network_interface",
	"azurerm_network_interface_security_group_association",
	"azurerm_network_security_group",
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

	// Azure Redis
	"azurerm_redis_firewall_rule",
	"azurerm_redis_linked_server",

	// Azure Registry
	"azurerm_container_registry_scope_map",
	"azurerm_container_registry_token",
	"azurerm_container_registry_webhook",

	// Azure SQL
	"azurerm_sql_server",

	// Azure Virtual Machines
	"azurerm_virtual_machine_data_disk_attachment",
}

var UsageOnlyResources []string = []string{}
