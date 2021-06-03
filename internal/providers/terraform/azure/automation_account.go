package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAutomationAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_account",
		RFunc: NewAzureRMAutomationAccount,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAutomationAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyJobRunMins, nonAzureConfigNodeCount, monthlyWatcherHours *decimal.Decimal
	group := d.References("resource_group_name")[0]
	location := strings.ToLower(group.Get("location").String())
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_job_run_mins").Type != gjson.Null {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(u.Get("monthly_job_run_mins").Int()))
		if monthlyJobRunMins.IsPositive() {
			costComponents = append(costComponents, RunTimeCostComponent(location, "500", "Basic Runtime", monthlyJobRunMins))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, RunTimeCostComponent(location, "500", "Basic Runtime", unknown))
	}

	if u != nil && u.Get("non_azure_config_node_count").Type != gjson.Null {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(u.Get("non_azure_config_node_count").Int()))
		if nonAzureConfigNodeCount.IsPositive() {
			costComponents = append(costComponents, NonNodesCostComponent(location, "5", "Non-Azure Node", nonAzureConfigNodeCount))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, NonNodesCostComponent(location, "5", "Non-Azure Node", unknown))
	}
	if u != nil && u.Get("monthly_watcher_hours").Type != gjson.Null {
		monthlyWatcherHours = decimalPtr(decimal.NewFromInt(u.Get("monthly_watcher_hours").Int()))
	}

	costComponents = append(costComponents, watchersCostComponent(location, "744", "Watcher", monthlyWatcherHours))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func watchersCostComponent(location, startUsage, meterName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Watchers",
		Unit:            "hours",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Automation"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
