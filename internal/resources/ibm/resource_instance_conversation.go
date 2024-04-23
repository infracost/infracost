package ibm

import (
	"fmt"
	"math"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const ENTERPRISE_MAU_PER_INSTANCE float64 = 50000
const ENTERPRISE_ADDITIONAL_PER_1K_USERS float64 = 1000
const PLUS_MAU_PER_INSTANCE float64 = 1000
const PLUS_ADDITIONAL_PER_100_USERS float64 = 100

/*
 * lite = "Lite" free plan
 * Trial = "Plus free trial plan"
 * Plus = "Plus pricing plan"
 * Enterprise == "Enterprise pricing plan"
 */
func GetWACostComponents(r *ResourceInstance) []*schema.CostComponent {
	if (r.Plan == "enterprise") || (r.Plan == "plus") {
		return []*schema.CostComponent{
			WAInstanceCostComponent(r),
			WAMonthlyActiveUsersCostComponent(r),
			WAMonthlyVoiceUsersCostComponent(r),
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
	} else if r.Plan == "plus-trial" {
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
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func WAInstanceCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	var name string
	if r.WA_Instance != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.WA_Instance))
	} else {
		q = decimalPtr(decimal.NewFromInt(1))
	}
	if r.Plan == "enterprise" {
		name = "Instance (50000 MAU included)"
	} else {
		name = "Instance (1000 MAU included)"
	}
	return &schema.CostComponent{
		Name:            name,
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

func WAMonthlyActiveUsersCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	var users_per_block float64
	var included_allotment float64
	var unit string

	if r.Plan == "enterprise" {
		included_allotment = ENTERPRISE_MAU_PER_INSTANCE
		users_per_block = ENTERPRISE_ADDITIONAL_PER_1K_USERS
		unit = "1K MAU"
	} else {
		included_allotment = PLUS_MAU_PER_INSTANCE
		users_per_block = PLUS_ADDITIONAL_PER_100_USERS
		unit = "100 MAU"
	}

	// if there are more active users than the monthly allotment of users included in the instance price, then create
	// a cost component for the additional users
	if r.WA_mau != nil {
		additional_users := *r.WA_mau - included_allotment
		if additional_users > 0 {
			// price for additional users charged is per 1k quantity, rounded up,
			// so 1001 additional users will equal 2 blocks of additional users
			q = decimalPtr(decimal.NewFromFloat(math.Ceil(additional_users / users_per_block)))
		}
	}

	return &schema.CostComponent{
		Name:            "Additional Monthly Active Users",
		Unit:            unit,
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
			Unit: strPtr("ACTIVE_USERS"),
		},
	}
}

func WAMonthlyVoiceUsersCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	var users_per_block float64
	var unit string

	if r.Plan == "enterprise" {
		users_per_block = ENTERPRISE_ADDITIONAL_PER_1K_USERS
		unit = "1K MAU"
	} else {
		users_per_block = PLUS_ADDITIONAL_PER_100_USERS
		unit = "100 MAU"
	}

	// price for voice users charged is per 1k quantity, rounded up,
	// so 1001 active users that used voice will equal 2 blocks of voice users
	if r.WA_vu != nil {
		voice_users := math.Ceil(*r.WA_vu / users_per_block)
		q = decimalPtr(decimal.NewFromFloat(voice_users))
	}

	return &schema.CostComponent{
		Name:            "Monthly Active Users using voice",
		Unit:            unit,
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
			Unit: strPtr("ACTIVE_VOICE_USERS"),
		},
	}
}
