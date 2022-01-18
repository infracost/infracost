package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeRouterNat struct {
	Address                *string
	Region                 *string
	AssignedVms            *int64   `infracost_usage:"assigned_vms"`
	MonthlyDataProcessedGb *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var ComputeRouterNatUsageSchema = []*schema.UsageItem{{Key: "assigned_vms", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *ComputeRouterNat) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeRouterNat) BuildResource() *schema.Resource {
	region := *r.Region

	var assignedVMs int64
	if r.AssignedVms != nil {
		assignedVMs = *r.AssignedVms
		if assignedVMs > 32 {
			assignedVMs = 32
		}
	}

	var dataProcessedGB *decimal.Decimal
	if r.MonthlyDataProcessedGb != nil {
		dataProcessedGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Assigned VMs (first 32)",
				Unit:           "VM-hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(assignedVMs)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/NAT Gateway: Uptime charge/")},
					},
				},
			},
			{
				Name:            "Data processed",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: dataProcessedGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/NAT Gateway: Data processing charge/")},
					},
				},
			},
		}, UsageSchema: ComputeRouterNatUsageSchema,
	}
}
