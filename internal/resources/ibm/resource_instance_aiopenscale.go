package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

/**
 * Lite: 'lite' (Free)
 * Essentials: 'essentials'
 * Standard V2 == 'standard-v2'
 */
func GetWGOVCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "essentials" {
		return []*schema.CostComponent{
			WGOVResourceUnitsCostComponent(r),
		}
	} else if r.Plan == "standard-v2" {
		return []*schema.CostComponent{
			WGOVModelCostComponent(r),
		}
	} else if r.Plan == "lite" {
		costComponent := &schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{costComponent}
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

/**
 * 1 RU for every:
 * 1 WGOV_PredictiveModelEvals
 * 1 WGOV_FoundationalModelEvals
 * 1 WGOV_GlobalExplanations
 * 500 WGOV_LocalExplanations
 * basically, everything converts into an RU for charging costs, with limits
 */
func WGOVResourceUnitsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WGOV_ru != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WGOV_ru))
	}
	return &schema.CostComponent{
		Name:            "Resource Units",
		Unit:            "RU",
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
			Unit: strPtr("RESOURCE_UNITS"),
		},
	}
}

/**
 * No restrictions with unlimited evaluations performed on a model, charged on a per model basis instead
 */
func WGOVModelCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WGOV_Models != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WGOV_Models))
	} else {
		q = decimalPtr(decimal.NewFromInt(1))
	}
	return &schema.CostComponent{
		Name:            "Deployed Models",
		Unit:            "Model",
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
			Unit: strPtr("MODELS_PER_MONTH"),
		},
	}
}
