package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
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

func (r *GlobalAccelerator) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent
	costComponent := &schema.CostComponent{
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
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.025)))
	costComponents = append(costComponents, costComponent)
	return &schema.Resource{
		Name:           r.Name,
		CostComponents: costComponents,
	}
}
