package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorAwsCloudTrailRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_aws_cloud_trail",
		RFunc: newSentinelDataConnectorAwsCloudTrail,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorAwsCloudTrail(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
