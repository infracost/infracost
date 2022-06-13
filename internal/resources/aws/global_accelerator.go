package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

// GlobalAccelerator struct represents AWS Global Accelerator service
//
// Resource information: https://aws.amazon.com/global-accelerator
// Pricing information: https://aws.amazon.com/global-accelerator/pricing/
type GlobalAccelerator struct {
	Name    string
	Address string
}

type globalAcceleratorRegionDataTransferUsage struct {
	US           *float64 `infracost_usage:"us"`
	Europe       *float64 `infracost_usage:"europe"`
	SouthAfrica  *float64 `infracost_usage:"south_africa"`
	SouthAmerica *float64 `infracost_usage:"south_america"`
	Japan        *float64 `infracost_usage:"japan"`
	Australia    *float64 `infracost_usage:"australia"`
	AsiaPacific  *float64 `infracost_usage:"asia_pacific"`
	India        *float64 `infracost_usage:"india"`
}

var (
	GlobalAcceleratorUsageSchema = []*schema.UsageItem{
		{
			Key:          "monthly_inbound_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_inbound_data_transfer_gb", Items: globalAcceleratorRegionDataTransferUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "monthly_outbound_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_outbound_data_transfer_gb", Items: globalAcceleratorRegionDataTransferUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
	}
	globalAcceleratorRegionDataTransferUsageSchema = []*schema.UsageItem{
		{Key: "us", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "europe", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "south_africa", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "south_america", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "japan", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "australia", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "asia_pacific", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "india", DefaultValue: 0, ValueType: schema.Int64},
	}
)

func (r *GlobalAccelerator) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *GlobalAccelerator) BuildResource() *schema.Resource {
	var (
		resource *schema.Resource = &schema.Resource{
			Name:           r.Name,
			UsageSchema:    GlobalAcceleratorUsageSchema,
			CostComponents: nil,
		}
	)

	costComponents := []*schema.CostComponent{
		r.fixedCostComponent(),
	}

	resource.CostComponents = costComponents
	return resource
}

func (r *GlobalAccelerator) fixedCostComponent() *schema.CostComponent {
	c := &schema.CostComponent{
		Name:           fmt.Sprintf("AWS Global Accelerator %s Fixed Fee", r.Name),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AWSGlobalAccelerator"),
		},
	}
	// AWS Global Accelerator has a fixed fee of 0.025$ per hour.
	// This price unfortunately is not mapped actually in AWS Pricing API
	// More: AWS_DEFAULT_REGION=us-east-1 aws pricing get-products --service-code AWSGlobalAccelerator | jq -r '.PriceList[] | fromjson | .product'
	c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.025)))
	return c
}
