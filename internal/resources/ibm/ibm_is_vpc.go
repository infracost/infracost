package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IbmIsVpc struct represents an instance of IBM's Virtual Private Cloud
//
//
// Resource information: https://registry.terraform.io/providers/IBM-Cloud/ibm/latest/docs/resources/is_vpc
// Pricing information: https://www.ibm.com/ca-en/cloud/vpc/pricing/
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

// @TODO: add/find a '$0' entry in pricing db to allow zero dollar cost components to show up in breakdown
func (r *IbmIsVpc) vpcCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.GigabyteTransmittedOutbounds != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.GigabyteTransmittedOutbounds))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPC egress %s", r.Region),
		Unit:            "Gigabyte",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpc"),
		},
	}
}

// @TODO add a cost component for monthly egress data in GB (traffic leaving the VPC)

// BuildResource builds a schema.Resource from a valid IbmIsVpc struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmIsVpc) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.vpcCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmIsVpcUsageSchema,
		CostComponents: costComponents,
	}
}
