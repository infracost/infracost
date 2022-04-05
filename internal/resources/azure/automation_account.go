package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type AutomationAccount struct {
	Address                 string
	Region                  string
	MonthlyJobRunMins       *int64 `infracost_usage:"monthly_job_run_mins"`
	NonAzureConfigNodeCount *int64 `infracost_usage:"non_azure_config_node_count"`
	MonthlyWatcherHours     *int64 `infracost_usage:"monthly_watcher_hours"`
}

var AutomationAccountUsageSchema = []*schema.UsageItem{{Key: "monthly_job_run_mins", ValueType: schema.Int64, DefaultValue: 0}, {Key: "non_azure_config_node_count", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_watcher_hours", ValueType: schema.Int64, DefaultValue: 0}}

func (r *AutomationAccount) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AutomationAccount) BuildResource() *schema.Resource {
	var monthlyJobRunMins, nonAzureConfigNodeCount, monthlyWatcherHours *decimal.Decimal
	location := r.Region
	costComponents := make([]*schema.CostComponent, 0)

	if r.MonthlyJobRunMins != nil {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(*r.MonthlyJobRunMins))
		if monthlyJobRunMins.IsPositive() {
			costComponents = append(costComponents, runTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
		}
	} else {
		costComponents = append(costComponents, runTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
	}

	if r.NonAzureConfigNodeCount != nil {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(*r.NonAzureConfigNodeCount))
		if nonAzureConfigNodeCount.IsPositive() {
			costComponents = append(costComponents, nonNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
		}
	} else {
		costComponents = append(costComponents, nonNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
	}
	if r.MonthlyWatcherHours != nil {
		monthlyWatcherHours = decimalPtr(decimal.NewFromInt(*r.MonthlyWatcherHours))
	}

	costComponents = append(costComponents, watchersCostComponent(location, "744", "Watcher", "Basic", monthlyWatcherHours))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: AutomationAccountUsageSchema,
	}
}

func watchersCostComponent(location, startUsage, meterName, skuName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Watchers",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Automation"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", skuName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
