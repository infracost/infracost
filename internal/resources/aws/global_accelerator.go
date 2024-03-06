package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// GlobalAccelerator struct represents AWS Global Accelerator service
//
// Resource information: https://aws.amazon.com/global-accelerator
// Pricing information: https://aws.amazon.com/global-accelerator/pricing/
type GlobalAccelerator struct {
	Address string
}

func (r *GlobalAccelerator) CoreType() string {
	return "FSxOpenZFSFileSystem"
}

func (r *GlobalAccelerator) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *GlobalAccelerator) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *GlobalAccelerator) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent
	costComponent := &schema.CostComponent{
		Name:           "Fixed fee",
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
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.025)))
	costComponents = append(costComponents, costComponent)
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
