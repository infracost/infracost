package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getSecurityCenterSubscriptionPricingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_security_center_subscription_pricing",
		CoreRFunc: newSecurityCenterSubscriptionPricing,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newSecurityCenterSubscriptionPricing(d *schema.ResourceData) schema.CoreResource {
	region := "Global"

	return &azure.SecurityCenterSubscriptionPricing{
		Address:      d.Address,
		Region:       region,
		Tier:         d.GetStringOrDefault("tier", "Free"),
		ResourceType: d.GetStringOrDefault("resource_type", "VirtualMachines"),
	}
}
