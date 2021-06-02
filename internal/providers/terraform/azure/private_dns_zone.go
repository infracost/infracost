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
	group := d.References("resource_group_name")
	location := group[0].Get("location").String()

	if strings.HasPrefix(strings.ToLower(location), "usgov") {
		location = "US Gov Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "germany") {
		location = "DE Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "china") {
		location = "Zone 1 (China)"
	}
	if location != "US Gov Zone 1" && location != "DE Zone 1" && location != "Zone 1 (China)" {
		location = "Zone 1"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hostedPublicZoneCostComponent(location))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
