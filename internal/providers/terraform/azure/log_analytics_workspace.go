package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

var sentinelDataConnectorRefs = []string{
	"azurerm_sentinel_data_connector_aws_cloud_trail.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_azure_active_directory.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_azure_advanced_threat_protection.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_azure_security_center.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_microsoft_cloud_app_security.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_office_365.log_analytics_workspace_id",
	"azurerm_sentinel_data_connector_threat_intelligence.log_analytics_workspace_id",
}

func getLogAnalyticsWorkspaceRegistryItem() *schema.RegistryItem {
	refs := []string{
		"resource_group_name",
		"azurerm_log_analytics_solution.workspace_resource_id",
	}

	return &schema.RegistryItem{
		Name:                "azurerm_log_analytics_workspace",
		CoreRFunc:           newLogAnalyticsWorkspace,
		ReferenceAttributes: append(refs, sentinelDataConnectorRefs...),
	}
}

func newLogAnalyticsWorkspace(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	sku := "PerGB2018"
	if !d.IsEmpty("sku") {
		sku = d.Get("sku").String()
	}

	sentinelEnabled := logAnalyticsWorkspaceDetectSentinel(d)

	capacity := d.Get("reservation_capacity_in_gb_per_day").Int()
	// Deprecated and removed in v3
	// this attribute typo was fixed in https://github.com/hashicorp/terraform-provider-azurerm/pull/14910
	// but we need to support the typo for backwards compatibility
	if !d.IsEmpty("reservation_capcity_in_gb_per_day") {
		capacity = d.Get("reservation_capcity_in_gb_per_day").Int()
	}

	return &azure.LogAnalyticsWorkspace{
		Address:                       d.Address,
		Region:                        region,
		SKU:                           sku,
		ReservationCapacityInGBPerDay: capacity,
		RetentionInDays:               d.Get("retention_in_days").Int(),
		SentinelEnabled:               sentinelEnabled,
	}
}

func logAnalyticsWorkspaceDetectSentinel(d *schema.ResourceData) bool {
	for _, ref := range sentinelDataConnectorRefs {
		if len(d.References(ref)) > 0 {
			return true
		}
	}

	logSolutionRefs := d.References("azurerm_log_analytics_solution.workspace_resource_id")
	if len(logSolutionRefs) > 0 {
		return logAnalyticsWorkspaceDetectSentinel(logSolutionRefs[0])
	}

	return false
}
