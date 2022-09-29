package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IbmIsVpc struct represents an instance of IBM's Virtual Private Cloud
//
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-getting-started
// Pricing information: https://www.ibm.com/cloud/vpc/pricing
type IbmIsVpc struct {
	Address string
	Region  string

	GigabyteTransmittedOutbounds *float64 `infracost_usage:"gigabyte_transmitted_outbounds"`
}

// IbmIsVpcUsageSchema defines a list which represents the usage schema of IbmIsVpc.
var IbmIsVpcUsageSchema = []*schema.UsageItem{
	{Key: "gigabyte_transmitted_outbounds", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IbmIsVpc.
// It uses the `infracost_usage` struct tags to populate data into the IbmIsVpc.
func (r *IbmIsVpc) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IbmIsVpc) vpcEgressFreeCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.GigabyteTransmittedOutbounds != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.GigabyteTransmittedOutbounds))
		if quantity.GreaterThan(decimal.NewFromInt(5)) {
			quantity = decimalPtr(decimal.NewFromInt(5))
		}
	}
	costComponent := schema.CostComponent{
		Name:            "VPC egress free allowance (first 5GB)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "nextgen-egress"))},
			},
		},
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &costComponent
}

func (r *IbmIsVpc) vpcEgressCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.GigabyteTransmittedOutbounds != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.GigabyteTransmittedOutbounds))
		if quantity.LessThanOrEqual(decimal.NewFromInt(5)) {
			quantity = decimalPtr(decimal.NewFromInt(0))
		} else {
			quantity = decimalPtr(quantity.Sub(decimal.NewFromInt(5)))
		}
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPC egress %s", r.Region),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpc"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "nextgen-egress"))},
			},
		},
	}
}

func (r *IbmIsVpc) vpcInstanceCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "VPC instance",
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpc"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IbmIsVpc struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmIsVpc) BuildResource() *schema.Resource {
	vpcCostComponent := r.vpcInstanceCostComponent()
	vpcCostComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	costComponents := []*schema.CostComponent{
		vpcCostComponent,
		r.vpcEgressFreeCostComponent(),
		r.vpcEgressCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmIsVpcUsageSchema,
		CostComponents: costComponents,
	}
}
