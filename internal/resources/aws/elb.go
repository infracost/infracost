package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ELB struct {
	Address                string
	Region                 string
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var ELBUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *ELB) CoreType() string {
	return "ELB"
}

func (r *ELB) UsageSchema() []*schema.UsageItem {
	return ELBUsageSchema
}

func (r *ELB) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ELB) BuildResource() *schema.Resource {
	var dataProcessed *decimal.Decimal
	if r.MonthlyDataProcessedGB != nil {
		dataProcessed = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.lbCostComponent(),
			r.dataProcessedCostComponent(dataProcessed),
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *ELB) lbCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Classic load balancer",
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr("Load Balancer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
			},
		},
	}
}

func (r *ELB) dataProcessedCostComponent(dataProcessed *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr("Load Balancer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/DataProcessing-Bytes/")},
			},
		},
		UsageBased: true,
	}
}
