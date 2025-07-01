package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type AutomationAccount struct {
	Address                 string
	Region                  string
	MonthlyJobRunMins       *int64 `infracost_usage:"monthly_job_run_mins"`
	NonAzureConfigNodeCount *int64 `infracost_usage:"non_azure_config_node_count"`
}

func (r *AutomationAccount) CoreType() string {
	return "AutomationAccount"
}

func (r *AutomationAccount) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_job_run_mins", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "non_azure_config_node_count", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *AutomationAccount) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AutomationAccount) BuildResource() *schema.Resource {
	var monthlyJobRunMins, nonAzureConfigNodeCount *decimal.Decimal
	location := r.Region
	costComponents := make([]*schema.CostComponent, 0)

	if r.MonthlyJobRunMins != nil {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(*r.MonthlyJobRunMins))
		if monthlyJobRunMins.IsPositive() {
			costComponents = append(costComponents, automationRunTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
		}
	} else {
		costComponents = append(costComponents, automationRunTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
	}

	if r.NonAzureConfigNodeCount != nil {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(*r.NonAzureConfigNodeCount))
		if nonAzureConfigNodeCount.IsPositive() {
			costComponents = append(costComponents, nonautomationDSCNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
		}
	} else {
		costComponents = append(costComponents, nonautomationDSCNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
