package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMLogAnalyticsWorkspaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_log_analytics_workspace",
		RFunc: newLogAnalyticsWorkspace,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newLogAnalyticsWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})
	sku := "PerGB2018"
	if !d.IsEmpty("sku") {
		sku = d.Get("sku").String()
	}

	capacity := d.Get("reservation_capacity_in_gb_per_day").Int()
	// this attribute typo was fixed in https://github.com/hashicorp/terraform-provider-azurerm/pull/14910
	// but we need to support the typo for backwards compatibility
	if !d.IsEmpty("reservation_capcity_in_gb_per_day") {
		capacity = d.Get("reservation_capcity_in_gb_per_day").Int()
	}

	r := &azure.LogAnalyticsWorkspace{
		Address:                       d.Address,
		Region:                        region,
		SKU:                           sku,
		ReservationCapacityInGBPerDay: capacity,
		RetentionInDays:               d.Get("retention_in_days").Int(),
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
