package azure

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMLinuxVirtualMachineRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources []string = []string{
	"azurerm_application_security_group",
	"azurerm_network_security_group",
	"azurerm_virtual_network",
}

var UsageOnlyResources []string = []string{}

// Other Notes:
// Only Basic Load Balancers are free of charge
