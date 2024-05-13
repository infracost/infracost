package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const SCC_STANDARD_PLAN_PROGRAMMATIC_NAME string = "security-compliance-center-standard-plan"
const SCC_TRIAL_PLAN_PROGRAMMATIC_NAME string = "security-compliance-center-trial-plan"

func GetSCCCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == SCC_STANDARD_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			SCCMonthlyEvaluationsCostComponent(r),
		}
	} else if r.Plan == SCC_TRIAL_PLAN_PROGRAMMATIC_NAME {
		costComponent := schema.CostComponent{
			Name:            "Trial",
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
			UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

/*
 * Evaluations:
 * - Standard: $USD/evaluation/month
 */
func SCCMonthlyEvaluationsCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	if r.SCC_Evaluations != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SCC_Evaluations)) // Quantity of current cost component (i.e. Number of evaluations performed in a month)
	}

	return &schema.CostComponent{
		Name:            "Evaluations",
		Unit:            "Evaluations",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("EVALUATION"),
		},
	}
}
