package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const CUH_PER_INSTANCE = 2500

/*
 * v2-professional = "Standard" pricing plan
 * v2-standard == "Essentials" pricing plan
 * lite = "Lite" free plan
 */
func GetWMLCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "v2-professional" {
		return []*schema.CostComponent{
			WMLInstanceCostComponent(r),
			WMLStandardCapacityUnitHoursCostComponent(r),
			WMLClass1ResourceUnitsCostComponent(r),
			WMLClass2ResourceUnitsCostComponent(r),
			WMLClass3ResourceUnitsCostComponent(r),
		}
	} else if r.Plan == "v2-standard" {
		return []*schema.CostComponent{
			WMLEssentialsCapacityUnitHoursCostComponent(r),
			WMLClass1ResourceUnitsCostComponent(r),
			WMLClass2ResourceUnitsCostComponent(r),
			WMLClass3ResourceUnitsCostComponent(r),
		}
	} else if r.Plan == "lite" {
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

func WMLInstanceCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WML_Instance != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WML_Instance))
	} else {
		q = decimalPtr(decimal.NewFromInt(1))
	}
	return &schema.CostComponent{
		Name:            "Instance (2500 CUH included)",
		Unit:            "Instance",
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
			Unit: strPtr("INSTANCES"),
		},
	}
}

func WMLEssentialsCapacityUnitHoursCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WML_CUH != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WML_CUH))
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

func WMLStandardCapacityUnitHoursCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	var instance float64

	if r.WML_Instance != nil {
		instance = *r.WML_Instance
	} else {
		instance = 1
	}
	if r.WML_CUH != nil {

		// standard plan is billed a fixed amount for each instance, which includes 2500 CUH's per instance.
		// if the used CUH exceeds the included quantity, the overage is charged at a flat rate.
		additional_cuh := *r.WML_CUH - (CUH_PER_INSTANCE * instance)
		if additional_cuh > 0 {
			q = decimalPtr(decimal.NewFromFloat(additional_cuh))
		}
	}

	return &schema.CostComponent{
		Name:            "Additional Capacity Unit-Hours",
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

func WMLClass1ResourceUnitsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WML_Class1RU != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WML_Class1RU))
	}
	return &schema.CostComponent{
		Name:            "Class 1 Resource Units",
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
			Unit: strPtr("CLASS_ONE_RESOURCE_UNITS"),
		},
	}
}

func WMLClass2ResourceUnitsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WML_Class1RU != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WML_Class2RU))
	}
	return &schema.CostComponent{
		Name:            "Class 2 Resource Units",
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
			Unit: strPtr("CLASS_TWO_RESOURCE_UNITS"),
		},
	}
}

func WMLClass3ResourceUnitsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.WML_Class1RU != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WML_Class3RU))
	}
	return &schema.CostComponent{
		Name:            "Class 3 Resource Units",
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
			Unit: strPtr("CLASS_THREE_RESOURCE_UNITS"),
		},
	}
}
