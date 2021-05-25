package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetStepFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sfn_state_machine",
		RFunc: NewStepFunction,
	}
}

func NewStepFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var durationMS, requests, transition, monthlyMemoryUsage, memoryRequest, memoryMB *decimal.Decimal
	memorySize := decimal.NewFromInt(64)
	tier := "STANDARD"
	region := d.Get("region").String()

	if d.Get("type").Type != gjson.Null {
		tier = d.Get("type").String()
	}
	costComponents := make([]*schema.CostComponent, 0)
	if tier == "STANDARD" {
		if u != nil && u.Get("monthly_transitions").Type != gjson.Null {
			transition = decimalPtr(decimal.NewFromInt(u.Get("monthly_transitions").Int()))
			transitionLimits := []int{4000}
			transitionQuantities := usage.CalculateTierBuckets(*transition, transitionLimits)
			if transition.LessThanOrEqual(decimal.NewFromInt(4000)) {
				return &schema.Resource{
					NoPrice:   true,
					IsSkipped: true,
				}
			}
			if transitionQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionStandardCostComponent("Transitions", region, tier, "0", &transitionQuantities[1]))
			}
		}
		if u == nil {
			costComponents = append(costComponents, stepFunctionStandardCostComponent("Transitions", region, tier, "0", transition))
		}
	}
	if tier == "EXPRESS" {
		if u != nil && u.Get("monthly_requests").Type != gjson.Null {
			requests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
			costComponents = append(costComponents, stepFunctionExpressRequestCostComponent("Requests", region, tier, requests))
		}
		if u != nil && u.Get("memory_mb").Type != gjson.Null {
			memoryRequest = decimalPtr(decimal.NewFromInt(u.Get("memory_mb").Int()))
			memoryMB = decimalPtr(calculateRequests(*memoryRequest, memorySize))
		}
		if u != nil && u.Get("workflow_duration_ms").Type != gjson.Null {
			durationMS = decimalPtr(decimal.NewFromInt(u.Get("workflow_duration_ms").Int()))
			monthlyMemoryUsage = decimalPtr(calculateGBSeconds(*memoryMB, *requests, *durationMS))
			pushLimits := []int{0, 3600000, 10800000}
			pushQuantities := usage.CalculateTierBuckets(*durationMS, pushLimits)
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (first 1K)", region, tier, "0", monthlyMemoryUsage))
			}
			if pushQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (next 4K)", region, tier, "3600000", monthlyMemoryUsage))
			}
			if pushQuantities[3].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (over 5K)", region, tier, "18000000", monthlyMemoryUsage))
			}
		}
		if u == nil {
			costComponents = append(costComponents, stepFunctionExpressRequestCostComponent("Requests", region, tier, requests))
			costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration", region, tier, "0", monthlyMemoryUsage))
		}

	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func stepFunctionStandardCostComponent(name, region, tier, startUsageAmt string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(1))))
	}
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1K transitions",
		UnitMultiplier:  1000,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonStates"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "servicename", Value: strPtr("AWS Step Functions")},
				{Key: "group", Value: strPtr("SFN-StateTransitions")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			StartUsageAmount: strPtr(startUsageAmt),
		},
	}
}

func stepFunctionExpressRequestCostComponent(name, region, tier string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(1))))
	}
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1M Requests",
		UnitMultiplier:  1000000,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonStates"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "servicename", Value: strPtr("AWS Step Functions")},
				{Key: "group", Value: strPtr("SFN-ExpressWorkflows-Requests")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func stepFunctionExpressDurationCostComponent(name, region, tier, startUsageAmt string, monthlyMemoryUsage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-Seconds",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyMemoryUsage,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonStates"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "servicename", Value: strPtr("AWS Step Functions")},
				{Key: "group", Value: strPtr("SFN-ExpressWorkflows-Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			StartUsageAmount: strPtr(startUsageAmt),
		},
	}
}
