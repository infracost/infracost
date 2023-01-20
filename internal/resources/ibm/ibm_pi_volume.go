package ibm

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IbmPiVolume struct represents a volume instance insider a Power Systems Virtual Server Workspace
//
// VolumePool, Type and AffinityPolicy=affinity are mutually exclusive
// AffinityInstance and AffinityVolume are required if AffinityPolicy=affinity
// If AffinityPolicy=affinity then the either the instance or the volume determine the type
//
// tier1=ssd, tier3=standard pricing wise
//
// Resource information: https://cloud.ibm.com/power/storage
// Pricing information: https://www.ibm.com/docs/en/power-systems-vs?topic=started-pricing-power-systems-virtual-servers
type IbmPiVolume struct {
	Address          string
	Region           string
	Name             string
	Size             int64
	Type             string // tier1, tier3, standard, ssd
	VolumePool       string
	AffinityPolicy   string // affinity, anti-affinity
	AffinityInstance string
	AffinityVolume   string

	MonthlyInstanceHours *float64 `infracost_usage:"monthly_instance_hours"`
}

// IbmPiVolumeUsageSchema defines a list which represents the usage schema of IbmPiVolume.
var IbmPiVolumeUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

var tierMapping = map[string]string{
	"tier1":    "TIER_ONE_STORAGE_GIGABYTE_HOURS",
	"ssd":      "TIER_ONE_STORAGE_GIGABYTE_HOURS",
	"tier3":    "TIER_THREE_STORAGE_GIGABYTE_HOURS",
	"standard": "TIER_THREE_STORAGE_GIGABYTE_HOURS",
}

// PopulateUsage parses the u schema.UsageData into the IbmPiVolume.
// It uses the `infracost_usage` struct tags to populate data into the IbmPiVolume.
func (r *IbmPiVolume) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid IbmPiVolume struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmPiVolume) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}
	var qInstanceHours *decimal.Decimal
	var qSize *decimal.Decimal
	var q *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		qInstanceHours = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
	}
	if r.Size != 0 {
		qSize = decimalPtr(decimal.NewFromInt(r.Size))
		if qInstanceHours != nil {
			q = decimalPtr(qInstanceHours.Mul(*qSize))
		}
	}
	if r.VolumePool != "" {
		// How am I supposed to guess?
		costComponent := &schema.CostComponent{
			Name:            "Price dependent upon volume pool settings",
			Unit:            "Gigabyte Instance Hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Region),
				Service:    strPtr("power-iaas"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: strPtr("power-virtual-server-group")},
				},
			},
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0.0)))
		costComponents = append(costComponents, costComponent)
	} else if r.AffinityPolicy == "affinity" {
		// How am I supposed to guess?
		costComponent := &schema.CostComponent{
			Name:            "Price dependent upon affinity settings",
			Unit:            "Gigabyte Instance Hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Region),
				Service:    strPtr("power-iaas"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: strPtr("power-virtual-server-group")},
				},
			},
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0.0)))
		costComponents = append(costComponents, costComponent)
	} else if r.Type == "tier1" || r.Type == "tier3" || r.Type == "standard" || r.Type == "ssd" {

		costComponent := &schema.CostComponent{
			Name:            "Price dependent upon affinity settings",
			Unit:            "Gigabyte Instance Hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Region),
				Service:    strPtr("power-iaas"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: strPtr("power-virtual-server-group")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr(tierMapping[r.Type]),
			},
		}
		costComponents = append(costComponents, costComponent)

	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmPiVolumeUsageSchema,
		CostComponents: costComponents,
	}
}
