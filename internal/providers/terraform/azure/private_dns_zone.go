package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNSPrivateZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_zone",
		RFunc: NewAzureRMDNSPrivateZone,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"Most expensive price tier is used."},
	}
}

func NewAzureRMDNSPrivateZone(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(region), "germany") {
		region = "DE Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(region), "china") {
		region = "Zone 1 (China)"
	}
	if region != "US Gov Zone 1" && region != "DE Zone 1" && region != "Zone 1 (China)" {
		region = "Zone 1"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hostedPublicZoneCostComponent(region))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
