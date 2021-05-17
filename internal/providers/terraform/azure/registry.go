package azure

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMAppServiceCertificateBindingRegistryItem(),
	GetAzureRMAppServiceCertificateOrderRegistryItem(),
	GetAzureRMLinuxVirtualMachineRegistryItem(),
	GetAzureRMLinuxVirtualMachineScaleSetRegistryItem(),
	GetAzureRMManagedDiskRegistryItem(),
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
	GetAzureRMPublicIPRegistryItem(),
	GetAzureRMPublicIPPrefixRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
	// Azure App Service
	"azurerm_app_service_virtual_network_swift_connection",

	// Azure Base
	"azurerm_resource_group",
	"azurerm_resource_provider_registration",
	"azurerm_subscription",
	"azurerm_role_assignment",

	// Azure Blueprints
	"azurerm_blueprint_assignment",

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
}

var UsageOnlyResources []string = []string{}
