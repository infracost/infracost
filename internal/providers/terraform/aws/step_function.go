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
	var storageGB, requests, transition *decimal.Decimal

	tier := "STANDARD"
	region := d.Get("region").String()

	if d.Get("type").Type != gjson.Null {
		tier = d.Get("type").String()
	}
	costComponents := make([]*schema.CostComponent, 0)
	if tier == "STANDARD" {
		if u != nil && u.Get("transition").Type != gjson.Null {
			transition = decimalPtr(decimal.NewFromInt(u.Get("transition").Int()))
			transitionLimits := []int{4000}
			transitionQuantities := usage.CalculateTierBuckets(*transition, transitionLimits)
			costComponents = append(costComponents, stepFunctionStandardCostComponent("State transition (Free tier (0-4000 transition))", region, tier, "free", &transitionQuantities[0]))
			if transitionQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionStandardCostComponent("State transition", region, tier, "0", &transitionQuantities[1]))
			}
		}
		if u == nil {
			costComponents = append(costComponents, stepFunctionStandardCostComponent("State transition", region, tier, "0", transition))
		}
	}
	if tier == "EXPRESS" {
		if u != nil && u.Get("requests").Type != gjson.Null {
			requests = decimalPtr(decimal.NewFromInt(u.Get("requests").Int()))
			costComponents = append(costComponents, stepFunctionExpressRequestCostComponent("Express workflow requests", region, tier, requests))
		}
		if u != nil && u.Get("storage_gb").Type != gjson.Null {
			storageGB = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
			pushLimits := []int{0, 3600000, 10800000}
			pushQuantities := usage.CalculateTierBuckets(*storageGB, pushLimits)
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Express workflow duration (0-1000 hours)", region, tier, "0", &pushQuantities[1]))
			}
			if pushQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Express workflow duration (1000-4000 hours)", region, tier, "3600000", &pushQuantities[2]))
			}
			if pushQuantities[3].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Express workflow duration (4000+ hours)", region, tier, "18000000", &pushQuantities[3]))
			}
		}
		if u == nil {
			costComponents = append(costComponents, stepFunctionExpressRequestCostComponent("Express workflow requests", region, tier, requests))
			costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Express workflow duration", region, tier, "0", storageGB))
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
		Unit:            "State transition",
		UnitMultiplier:  1,
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
		Unit:            "Requests",
		UnitMultiplier:  1,
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

func stepFunctionExpressDurationCostComponent(name, region, tier, startUsageAmt string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(3600))))
	}
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-Seconds",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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
