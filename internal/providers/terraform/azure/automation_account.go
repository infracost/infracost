package azure

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMAutomationAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_account",
		RFunc: NewAzureRMAutomationAccount,
	}
}

func NewAzureRMAutomationAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyJobRunMins, nonAzureConfigNodeCount, monthlyWatcherHours *decimal.Decimal
	location := lookupRegion(d, []string{})
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_job_run_mins").Type != gjson.Null {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(u.Get("monthly_job_run_mins").Int()))
		if monthlyJobRunMins.IsPositive() {
			costComponents = append(costComponents, runTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
		}
	} else {
		costComponents = append(costComponents, runTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))
	}

	if u != nil && u.Get("non_azure_config_node_count").Type != gjson.Null {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(u.Get("non_azure_config_node_count").Int()))
		if nonAzureConfigNodeCount.IsPositive() {
			costComponents = append(costComponents, nonNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
		}
	} else {
		costComponents = append(costComponents, nonNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))
	}
	if u != nil && u.Get("monthly_watcher_hours").Type != gjson.Null {
		monthlyWatcherHours = decimalPtr(decimal.NewFromInt(u.Get("monthly_watcher_hours").Int()))
	}

	costComponents = append(costComponents, watchersCostComponent(location, "744", "Watcher", "Basic", monthlyWatcherHours))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
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
