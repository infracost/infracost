package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// GlobalAccelerator struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/<PATH/TO/RESOURCE>/
// Pricing information: https://aws.amazon.com/<PATH/TO/PRICING>/
type GlobalAccelerator struct {
	Name          string
	IPAddressType string
	Enabled       bool

	// "usage" args
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var GlobalAcceleratorUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
}

func (r *GlobalAccelerator) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *GlobalAccelerator) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.fixedCostComponent(),
		// Below is an example of a cost component built with the parsed usage property.
		// Note the r.MonthlyDataProcessedGB field passed to hourly quantity.
		// r.dataTransferCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Name,
		UsageSchema:    GlobalAcceleratorUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *GlobalAccelerator) fixedCostComponent() *schema.CostComponent {
	c := &schema.CostComponent{
		Name:           "Global Accelerator",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		MonthlyCost:    decimalPtr(decimal.NewFromFloat(18)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AWSGlobalAccelerator"),
		},
	}
	// AWS Global Accelerator has a fixed fee of 0.025$ per hour.
	// This price unfortunately is not mapped in AWS Pricing API
	// More: AWS_DEFAULT_REGION=us-east-1 aws pricing describe-services | jq -r '.PriceList[] | fromjson | .product'
	c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.025)))
	return c
}

func (r *GlobalAccelerator) dataTransferCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Data processed",
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataProcessedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AWSGlobalAccelerator"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("UsageBytes")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
