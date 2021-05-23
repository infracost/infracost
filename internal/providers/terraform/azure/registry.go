package azure

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMAppServiceCertificateBindingRegistryItem(),
	GetAzureRMAppServiceCertificateOrderRegistryItem(),
	GetAzureRMDatabricksWorkspaceRegistryItem(),
	GetAzureRMFirewallRegistryItem(),
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
	GetAzureRMNotificationHubsRegistryItem(),
	GetAzureRMEventHubsRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
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

	// Azure Event Hub
	"azurerm_eventhub",
}

var UsageOnlyResources []string = []string{}
