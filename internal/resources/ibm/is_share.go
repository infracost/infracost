package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsShare struct represents: A floating IP instance inside a VPC
//
// Resource information: https://registry.terraform.io/providers/IBM-Cloud/ibm/latest/docs/resources/is_share
// Pricing information: https://www.ibm.com/cloud/vpc/pricing
type IsShare struct {
	Address              string
	Region               string
	Zone                 string
	Profile              string // general-purpose, 5iops-tier, 10iops-tier, custom, dp2 (plan for zonal file storage. tiered profiles are deprecated)
	Size                 int64
	IOPS                 int64 // optional maximum input/output per second, limited by size.
	IsReplica            bool
	InlineReplicaName    string
	InlineReplicaZone    string
	MonthlyInstanceHours *float64 `infracost_usage:"is-share_monthly_instance_hours"`
	TransmittedGB        *int64   `infracost_usage:"is-share_monthly_transmitted_gb"`
}

// IsShareUsageSchema defines a list which represents the usage schema of IsShare.
var IsShareUsageSchema = []*schema.UsageItem{
	{Key: "is-share_monthly_instance_hours", DefaultValue: 730, ValueType: schema.Float64},
	{Key: "is-share_monthly_transmitted_gb", DefaultValue: 0, ValueType: schema.Int64},
}

// PopulateUsage parses the u schema.UsageData into the IsShare.
// It uses the `infracost_usage` struct tags to populate data into the IsShare.
func (r *IsShare) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsShare) isShareCostComponent(isReplica bool) []*schema.CostComponent {
	maxIops := r.IOPS
	if maxIops == 0 {
		maxIops = 100
	}
	var gigabyteHoursQ *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		gigabyteHoursQ = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
		gigabyteHoursQ = decimalPtr(gigabyteHoursQ.Mul(decimal.NewFromInt(r.Size)))
	}
	var iopsHoursQ *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		iopsHoursQ = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
		iopsHoursQ = decimalPtr(iopsHoursQ.Mul(decimal.NewFromInt(maxIops)))
	}

	zone := r.Zone
	shareType := ""
	if isReplica {
		zone = r.InlineReplicaZone
		shareType = "Replica "
	}

	if r.IsReplica {
		shareType = "Replica "
	}

	return []*schema.CostComponent{
		{
			Name:            fmt.Sprintf("%d GB %sfile storage share for VPC - %s", r.Size, shareType, zone),
			Unit:            "GB Hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: gigabyteHoursQ,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("ibm"),
				Region:        strPtr(r.Region),
				Service:       strPtr("is.share"),
				ProductFamily: strPtr("service"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: &r.Profile},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("GIGABYTE_HOURS"),
			},
		},
		{ // graduated tier pricing for IOPS Hours (quantity in that tier x unit price at that tier + quantity in next tier x unit price at that tier + ...)
			Name:            fmt.Sprintf("%d Max IOPS", maxIops),
			Unit:            "IOPS Hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: iopsHoursQ,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("ibm"),
				Region:        strPtr(r.Region),
				Service:       strPtr("is.share"),
				ProductFamily: strPtr("service"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: &r.Profile},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("IOPS_HOURS"),
			},
		},
	}
}

func (r *IsShare) transmittedDataCostComponent() *schema.CostComponent {
	var transmittedQ *decimal.Decimal
	if r.TransmittedGB != nil {
		transmittedQ = decimalPtr(decimal.NewFromInt(*r.TransmittedGB))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Transmitted data"),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: transmittedQ,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("is.share"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Profile},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTEDS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsShare struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsShare) BuildResource() *schema.Resource {
	costComponents := r.isShareCostComponent(false)
	if r.IsReplica {
		costComponents = append(costComponents, r.transmittedDataCostComponent())
	}
	// add a cost component for another share if there's an inline replica share found
	if r.InlineReplicaName != "" {
		costComponents = append(costComponents, r.isShareCostComponent(true)...)
		costComponents = append(costComponents, r.transmittedDataCostComponent())
	}
	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsShareUsageSchema,
		CostComponents: costComponents,
	}
}
