package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsVolume struct represents a block storage attached to VPC
//
// Catalog link: https://cloud.ibm.com/vpc-ext/provision/storage
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-creating-block-storage&interface=ui
// Pricing information: https://www.ibm.com/cloud/vpc/pricing
type IsVolume struct {
	Address  string
	Region   string
	Profile  string // general-purpose, 5iops-tier, 10iops-tier, custom
	Capacity int64
	// Only for custom profile
	IOPS int64

	MonthlyInstanceHours *float64 `infracost_usage:"monthly_instance_hours"`
}

// IsVolumeUsageSchema defines a list which represents the usage schema of IsVolume.
var IsVolumeUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsVolume.
// It uses the `infracost_usage` struct tags to populate data into the IsVolume.
func (r *IsVolume) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsVolume) CustomCostComponenets() []*schema.CostComponent {
	var gigabyteHoursQ *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		gigabyteHoursQ = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
		gigabyteHoursQ = decimalPtr(gigabyteHoursQ.Mul(decimal.NewFromInt(r.Capacity)))
	}
	var iopsHoursQ *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		iopsHoursQ = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
		iopsHoursQ = decimalPtr(iopsHoursQ.Mul(decimal.NewFromInt(r.IOPS)))
	}
	return []*schema.CostComponent{
		{
			Name:            "Gigabyte Hours",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: gigabyteHoursQ,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("ibm"),
				ProductFamily: strPtr("service"),
				Service:       strPtr("is.volume"),
				Region:        strPtr(r.Region),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", ValueRegex: regexPtr(fmt.Sprintf("gen2-volume-%s", r.Profile))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("GIGABYTE_HOURS"),
			},
		},
		{
			Name:            "IOPS Hours",
			Unit:            "IOPS",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: iopsHoursQ,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("ibm"),
				ProductFamily: strPtr("service"),
				Service:       strPtr("is.volume"),
				Region:        strPtr(r.Region),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", ValueRegex: regexPtr(fmt.Sprintf("gen2-volume-%s", r.Profile))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("IOPS_HOURS"),
			},
		},
	}
}

func (r *IsVolume) PlannedCostComponenets() []*schema.CostComponent {
	var q *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
		q = decimalPtr(q.Mul(decimal.NewFromInt(r.Capacity)))
	}
	return []*schema.CostComponent{
		{
			Name:            "Gigabyte Hours",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("ibm"),
				ProductFamily: strPtr("service"),
				Service:       strPtr("is.volume"),
				Region:        strPtr(r.Region),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", ValueRegex: regexPtr(fmt.Sprintf("gen2-volume-%s", r.Profile))},
				},
			},
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsVolume struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsVolume) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.Profile == "custom" {
		costComponents = r.CustomCostComponenets()
	} else {
		costComponents = r.PlannedCostComponenets()
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsVolumeUsageSchema,
		CostComponents: costComponents,
	}
}
