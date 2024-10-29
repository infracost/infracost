package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type AutomationJobSchedule struct {
	Address           string
	Region            string
	MonthlyJobRunMins *int64 `infracost_usage:"monthly_job_run_mins"`
}

func (r *AutomationJobSchedule) CoreType() string {
	return "AutomationJobSchedule"
}

func (r *AutomationJobSchedule) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_job_run_mins", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *AutomationJobSchedule) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AutomationJobSchedule) BuildResource() *schema.Resource {
	var monthlyJobRunMins *decimal.Decimal
	location := r.Region

	if r.MonthlyJobRunMins != nil {
		monthlyJobRunMins = decimalPtr(decimal.NewFromInt(*r.MonthlyJobRunMins))
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, automationRunTimeCostComponent(location, "500", "Basic Runtime", "Basic", monthlyJobRunMins))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func automationRunTimeCostComponent(location, startUsage, meterName, skuName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
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
		UsageBased: true,
	}
}
