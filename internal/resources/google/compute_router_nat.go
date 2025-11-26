package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeRouterNAT struct {
	Address                string
	Region                 string
	AssignedVMs            *int64   `infracost_usage:"assigned_vms"`
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *ComputeRouterNAT) CoreType() string {
	return "ComputeRouterNAT"
}

func (r *ComputeRouterNAT) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "assigned_vms", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *ComputeRouterNAT) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeRouterNAT) BuildResource() *schema.Resource {
	var assignedVMs int64
	if r.AssignedVMs != nil {
		assignedVMs = min(*r.AssignedVMs, 32)
	}

	var dataProcessedGB *decimal.Decimal
	if r.MonthlyDataProcessedGB != nil {
		dataProcessedGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Assigned VMs (first 32)",
				Unit:           "VM-hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(assignedVMs)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/NAT Gateway: Uptime charge/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Data processed",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: dataProcessedGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/NAT Gateway: Data processing charge/")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
