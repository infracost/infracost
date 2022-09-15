package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsPublicGateway struct a Public Gateway for a VPC
//
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-public-gateways
// Pricing information: https://www.ibm.com/cloud/virtual-servers/pricing
type IsPublicGateway struct {
	Address string
	Region  string
}

// IsPublicGatewayUsageSchema defines a list which represents the usage schema of IsPublicGateway.
var IsPublicGatewayUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the IsPublicGateway.
// It uses the `infracost_usage` struct tags to populate data into the IsPublicGateway.
func (r *IsPublicGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsPublicGateway) publicGatewayCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Public gateway Floating IP %s", r.Region),
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("is.floating-ip"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("default-nextgen")},
			},
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsPublicGateway struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsPublicGateway) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.publicGatewayCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsPublicGatewayUsageSchema,
		CostComponents: costComponents,
	}
}
