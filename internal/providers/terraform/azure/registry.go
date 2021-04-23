package azure

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMAppServiceCertificateOrderRegistryItem(),
	GetAzureRMLinuxVirtualMachineRegistryItem(),
	GetAzureRMLinuxVirtualMachineScaleSetRegistryItem(),
	GetAzureRMManagedDiskRegistryItem(),
	GetAzureMariaDBServerRegistryItem(),
	GetAzureMySQLServerRegistryItem(),
	GetAzurePostgreSQLServerRegistryItem(),
	GetAzureRMWindowsVirtualMachineRegistryItem(),
	GetAzureRMWindowsVirtualMachineScaleSetRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
	// Azure Base
	"azurerm_resource_group",
	"azurerm_resource_provider_registration",
	"azurerm_subscription",

	// Azure Blueprints
	"azurerm_blueprint_assignment",

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
