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
	GetAzureRMWindowsVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineScaleSetRegistryItem(),
	GetAzureRMAppServicePlanRegistryItem(),
	GetAzureRMAppIsolatedServicePlanRegistryItem(),
	GetAzureRMContainerRegistryRegistryItem(),
	GetAzureRMAppFunctionRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
	// Azure Base
	"azurerm_resource_group",
	"azurerm_resource_provider_registration",
	"azurerm_subscription",

	// Azure Blueprints
	"azurerm_blueprint_assignment",
	
	// Azure ContainerRegistry
	"azurerm_container_registry_scope_map",
	"azurerm_container_registry_token",
	"azurerm_container_registry_webhook",

	// Azure Container Registry
	"azurerm_container_registry_scope_map",
	"azurerm_container_registry_token",
	"azurerm_container_registry_webhook",

	// Azure Networking
	"azurerm_application_security_group",
	"azurerm_network_security_group",
	"azurerm_virtual_network",

	// Azure Policy
	"azurerm_policy_assignment",
	"azurerm_policy_definition",
	"azurerm_policy_remediation",
	"azurerm_policy_set_definition",
}

var UsageOnlyResources []string = []string{}

// Other Notes:
// Only Basic Load Balancers are free of charge
