package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getLogAnalyticsSolutionRegistryItem() *schema.RegistryItem {
	refs := []string{
		"resource_group_name",
		"workspace_resource_id",
	}

	return &schema.RegistryItem{
		Name:                "azurerm_log_analytics_solution",
		RFunc:               newLogAnalyticsSolution,
		ReferenceAttributes: append(refs, sentinelDataConnectorRefs...),
	}
}

func newLogAnalyticsSolution(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
