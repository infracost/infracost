package azure

import (
	"fmt"
	"strings"

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

	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: JobRunTimeCostComponent(d, u, group),
	}
}

func JobRunTimeCostComponent(d *schema.ResourceData, u *schema.UsageData, group *schema.ResourceData) []*schema.CostComponent {
	var monthlyJobRunMins *decimal.Decimal

	location := strings.ToLower(group.Get("location").String())
	if u != nil && u.Get("monthly_job_run_mins").Type != gjson.Null {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(u.Get("monthly_job_run_mins").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, RunTimeCostComponent(location, "500", "Basic Runtime", monthlyJobRunMins))

	return costComponents
}

func RunTimeCostComponent(location, startUsage, meterName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Job run time",
		Unit:            "minutes",
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
