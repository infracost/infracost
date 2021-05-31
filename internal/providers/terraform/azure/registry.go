package azure

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMApiManagementRegistryItem(),
	GetAzureRMAppServiceCertificateBindingRegistryItem(),
	GetAzureRMAppServiceCertificateOrderRegistryItem(),
	GetAzureRMCDNEndpointRegistryItem(),
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
	GetAzureRMFirewallRegistryItem(),
	GetAzureRMHDInsightHadoopClusterRegistryItem(),
	GetAzureRMHDInsightHBaseClusterRegistryItem(),
	GetAzureRMHDInsightInteractiveQueryClusterRegistryItem(),
	GetAzureRMHDInsightKafkaClusterRegistryItem(),
	GetAzureRMHDInsightSparkClusterRegistryItem(),
	GetAzureRMLinuxVirtualMachineRegistryItem(),
	GetAzureRMLinuxVirtualMachineScaleSetRegistryItem(),
	GetAzureRMManagedDiskRegistryItem(),
	GetAzureRMKeyVaultCertificateRegistryItem(),
	GetAzureRMKeyVaultKeyRegistryItem(),
	GetAzureRMKeyVaultManagedHSMRegistryItem(),
	GetAzureRMKubernetesClusterRegistryItem(),
	GetAzureRMKubernetesClusterNodePoolRegistryItem(),
	GetAzureMariaDBServerRegistryItem(),
	GetAzureMSSQLDatabaseRegistryItem(),
	GetAzureMySQLServerRegistryItem(),
	GetAzurePostgreSQLServerRegistryItem(),
	GetAzureStorageAccountRegistryItem(),
	GetAzureRMVirtualMachineScaleSetRegistryItem(),
	GetAzureRMVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineScaleSetRegistryItem(),
	GetAzureRMAppServicePlanRegistryItem(),
	GetAzureRMAppIsolatedServicePlanRegistryItem(),
	GetAzureRMAppFunctionRegistryItem(),
	GetAzureRMContainerRegistryRegistryItem(),
	GetAzureRMAppIntegrationServiceEnvironmentRegistryItem(),
	GetAzureRMPublicIPRegistryItem(),
	GetAzureRMPublicIPPrefixRegistryItem(),
	GetAzureRMAppNATGatewayRegistryItem(),
	GetAzureRMAppServiceCustomHostnameBindingRegistryItem(),
	GetAzureRMNotificationHubNamespaceRegistryItem(),
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

	// Azure Networking
	"azurerm_application_security_group",
	"azurerm_network_interface",
	"azurerm_network_interface_security_group_association",
	"azurerm_network_security_group",
	"azurerm_subnet",
	"azurerm_subnet_network_security_group_association",
	"azurerm_virtual_network",

	// Azure Policy
	"azurerm_policy_assignment",
	"azurerm_policy_definition",
	"azurerm_policy_remediation",
	"azurerm_policy_set_definition",

	// Azure Registry
	"azurerm_container_registry_scope_map",
	"azurerm_container_registry_token",
	"azurerm_container_registry_webhook",

	// Azure Virtual Machines
	"azurerm_virtual_machine_data_disk_attachment",

	// Azure Notification Hub
	"azurerm_notification_hub",
}

var UsageOnlyResources []string = []string{}
