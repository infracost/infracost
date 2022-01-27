package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type NeptuneClusterInstance struct {
	Address             string
	Region              string
	InstanceClass       string
	Count               *int64
	MonthlyCPUCreditHrs *int64 `infracost_usage:"monthly_cpu_credit_hrs"`
}

var NeptuneClusterInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
}

func (r *NeptuneClusterInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneClusterInstance) BuildResource() *schema.Resource {
	hourlyQuantity := 1
	if r.Count != nil {
		hourlyQuantity = int(*r.Count)
	}

	var monthlyCPUCreditHrs *decimal.Decimal
	if r.MonthlyCPUCreditHrs != nil {
		monthlyCPUCreditHrs = decimalPtr(decimal.NewFromInt(*r.MonthlyCPUCreditHrs))
	}

	costComponents := []*schema.CostComponent{
		r.dbInstanceCostComponent(hourlyQuantity),
	}

	if strings.HasPrefix(strings.ToLower(r.InstanceClass), "db.t3.") {
		costComponents = append(costComponents, r.cpuCreditsCostComponent(monthlyCPUCreditHrs))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    NeptuneClusterInstanceUsageSchema,
	}
}

func (r *NeptuneClusterInstance) dbInstanceCostComponent(quantity int) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Database instance (on-demand, %s)", r.InstanceClass),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(quantity))),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", r.InstanceClass))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *NeptuneClusterInstance) cpuCreditsCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "CPU credits",
		Unit:           "vCPU-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("APE1-CPUCredits:db.t3")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
