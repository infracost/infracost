package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAutomationJobScheduleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_job_schedule",
		RFunc: NewAzureRMAutomationJobSchedule,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAutomationJobSchedule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyJobRunMins *decimal.Decimal
	location := lookupRegion(d, []string{"resource_group_name"})

	if u != nil && u.Get("monthly_job_run_mins").Type != gjson.Null {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(u.Get("monthly_job_run_mins").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, runTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func runTimeCostComponent(location, startUsage, meterName, skuName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Job run time",
		Unit:            "minutes",
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
