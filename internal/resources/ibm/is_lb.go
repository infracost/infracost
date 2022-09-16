package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsLb a load balanacer for a VPC
//
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-nlb-vs-elb
// Pricing information: https://www.ibm.com/cloud/virtual-servers/pricing
type IsLb struct {
	Address string
	Region  string
	Logging bool
	Profile string // "network" or "application"
	Type    string // "public" or "private"

	MonthlyInstanceHours *float64 `infracost_usage:"monthly_instance_hours"`
	GigabyteProcessed    *float64 `infracost_usage:"gigabyte_processed"`
}

// IsLbUsageSchema defines a list which represents the usage schema of IsLb.
var IsLbUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "gigabyte_processed", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsLb.
// It uses the `infracost_usage` struct tags to populate data into the IsLb.
func (r *IsLb) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsLb) instanceHoursCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
	}

	var planName string
	if r.Profile == "application" {
		planName = "gen2-load-balancer"
	} else {
		planName = "network-load-balancer-gen2"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance hours (%s, %s)", r.Region, r.Profile),
		Unit:            "Instance hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.load-balancer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr(planName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCE_HOURS"),
		},
	}
}

func (r *IsLb) gigabyteProcessedCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.GigabyteProcessed != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.GigabyteProcessed))
	}

	var planName string
	if r.Profile == "application" {
		planName = "gen2-load-balancer"
	} else {
		planName = "network-load-balancer-gen2"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Data processed (%s , %s)", r.Region, r.Profile),
		Unit:            "GB months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.load-balancer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr(planName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_MONTHS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsLb struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsLb) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.instanceHoursCostComponent(),
		r.gigabyteProcessedCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsLbUsageSchema,
		CostComponents: costComponents,
	}
}
