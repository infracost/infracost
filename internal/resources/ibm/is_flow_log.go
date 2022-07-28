package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsFlowLog struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.ibm.com/<PATH/TO/RESOURCE>/
// Pricing information: https://cloud.ibm.com/<PATH/TO/PRICING>/
type IsFlowLog struct {
	Address       string
	Region        string
	TransmittedGB *int64 `infracost_usage:"transmitted_gb"`
}

// IsFlowLogUsageSchema defines a list which represents the usage schema of IsFlowLog.
var IsFlowLogUsageSchema = []*schema.UsageItem{
	{Key: "transmitted_gb", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsFlowLog.
// It uses the `infracost_usage` struct tags to populate data into the IsFlowLog.
func (r *IsFlowLog) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsFlowLog) isFlowLogFreeCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.TransmittedGB != nil {
		q = decimalPtr(decimal.NewFromInt(*r.TransmittedGB))
		if q.GreaterThan(decimal.NewFromInt(5)) {
			q = decimalPtr(decimal.NewFromInt(5))
		}
	}
	costComponent := schema.CostComponent{
		Name:            fmt.Sprintf("Flow Log Collector %s - free allowance (first 5GB)", r.Region),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("is.flow-log-collector"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("flow-log-standard-paid-plan")},
				{Key: "region", Value: strPtr(r.Region)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTEDS"),
		},
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &costComponent
}

func (r *IsFlowLog) isFlowLogCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.TransmittedGB != nil {
		q = decimalPtr(decimal.NewFromInt(*r.TransmittedGB))
		if q.LessThanOrEqual(decimal.NewFromInt(5)) {
			q = decimalPtr(decimal.NewFromInt(0))
		} else {
			q = decimalPtr(q.Sub(decimal.NewFromInt(5)))
		}
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Flow Log Collector %s", r.Region),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("is.flow-log-collector"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("flow-log-standard-paid-plan")},
				{Key: "region", Value: strPtr(r.Region)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTEDS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsFlowLog struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsFlowLog) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.isFlowLogFreeCostComponent(),
		r.isFlowLogCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsFlowLogUsageSchema,
		CostComponents: costComponents,
	}
}
