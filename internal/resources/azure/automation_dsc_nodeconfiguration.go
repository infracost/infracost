package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type AutomationDSCNodeConfiguration struct {
	Address string
	Region  string

	NonAzureConfigNodeCount *int64 `infracost_usage:"non_azure_config_node_count"`
}

var AutomationDSCNodeConfigurationUsageSchema = []*schema.UsageItem{{Key: "non_azure_config_node_count", ValueType: schema.Int64, DefaultValue: 0}}

func (r *AutomationDSCNodeConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AutomationDSCNodeConfiguration) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: automationDSCNodesCostComponent(&r.Region, r.NonAzureConfigNodeCount),
		UsageSchema:    AutomationDSCNodeConfigurationUsageSchema,
	}
}
