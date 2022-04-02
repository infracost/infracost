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
			"azurerm_sentinel_data_connector_threat_intelligence.log_analytics_workspace_id",
			"azurerm_sentinel_data_connector_office_365.log_analytics_workspace_id",
			"azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection.log_analytics_workspace_id",
			"azurerm_sentinel_data_connector_microsoft_cloud_app_security",
			"azurerm_sentinel_data_connector_azure_security_center",
			"azurerm_sentinel_data_connector_azure_advanced_threat_protection",
			"azurerm_sentinel_data_connector_azure_active_directory",
			"azurerm_sentinel_data_connector_aws_cloud_trail",
		},
	}
}

func newLogAnalyticsWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})
	sku := "PerGB2018"
	if !d.IsEmpty("sku") {
		sku = d.Get("sku").String()
	}

	sentinelRefs := [][]*schema.ResourceData{
		d.References("azurerm_sentinel_data_connector_threat_intelligence.log_analytics_workspace_id"),
		d.References("azurerm_sentinel_data_connector_office_365.log_analytics_workspace_id"),
		d.References("azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection.log_analytics_workspace_id"),
		//...
	}

	sentinelEnabled := false
	for _, ref := range sentinelRefs {
		if len(ref) > 0 {
			sentinelEnabled = true
			break
		}
	}

	capacity := d.Get("reservation_capacity_in_gb_per_day").Int()
	// Deprecated and removed in v3
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
		SentinelEnabled:               sentinelEnabled,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
