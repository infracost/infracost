package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

/*
 * professional-v1 = "Professional" pricing plan
 * free-v1 = "Lite" free plan
 */
func GetWSCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "professional-v1" {
		return []*schema.CostComponent{
			WSCapacityUnitHoursCostComponent(r),
		}
	} else if r.Plan == "free-v1" {
		costComponent := schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func WSCapacityUnitHoursCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WS_CUH != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WS_CUH))
	}
	return &schema.CostComponent{
		Name:            "Capacity Unit-Hours",
		Unit:            "CUH",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CAPACITY_UNIT_HOURS"),
		},
	}
}
