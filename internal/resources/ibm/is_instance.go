package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsInstance struct represents an IBM virtual server instance.
//
// Pricing information: https://cloud.ibm.com/kubernetes/catalog/about

type IsInstance struct {
	Address     string
	Region      string
	Profile     string // should be values from CLI 'ibmcloud is instance-profiles'
	Zone        string
	IsDedicated bool // will be true if a dedicated_host or dedicated_host_group is specified
	BootVolume  struct {
		Name string
		Size int64
	}
	MonthlyInstanceHours *float64 `infracost_usage:"monthly_instance_hours"`
}

var IsInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsInstance.
// It uses the `infracost_usage` struct tags to populate data into the IsInstance.
func (r *IsInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsInstance) instanceHoursCostComponent() *schema.CostComponent {

	unit := "RESERVATION_HOURS_HOURLY"
	planName := fmt.Sprintf("instance-%s", r.Profile)
	unitMultiplier := int64(1)
	var q *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
	}
	if r.IsDedicated {
		q = decimalPtr(decimal.NewFromFloat(0))
		unitMultiplier = 0
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance Hours (%s)", r.Profile),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(unitMultiplier),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("is.reservation"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &planName},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func (r *IsInstance) bootVolumeCostComponent() *schema.CostComponent {

	var q *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		q = decimalPtr(decimal.NewFromFloat(float64(r.BootVolume.Size) * (*r.MonthlyInstanceHours)))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Boot volume (%s, %d GB)", r.BootVolume.Name, r.BootVolume.Size),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			ProductFamily: strPtr("service"),
			Service:       strPtr("is.volume"),
			Region:        strPtr(r.Region),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: regexPtr(("gen2-volume-general-purpose"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_HOURS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsShare struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsInstance) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.instanceHoursCostComponent(),
		r.bootVolumeCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsInstanceUsageSchema,
		CostComponents: costComponents,
	}
}
