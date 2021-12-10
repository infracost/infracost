package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/config"
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

func NewStepFunction(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	var duration, memoryRequest, requests, transitions, gbSeconds *decimal.Decimal

	tier := "STANDARD"
	if d.Get("type").Type != gjson.Null {
		tier = d.Get("type").String()
	}

	if strings.ToLower(tier) == "standard" {
		if u != nil && u.Get("monthly_transitions").Type != gjson.Null {
			transitions = decimalPtr(decimal.NewFromInt(u.Get("monthly_transitions").Int()))
		}
		costComponents = append(costComponents, stepFunctionStandardCostComponent(region, transitions))
	}

	if strings.ToLower(tier) == "express" {
		if u != nil && u.Get("monthly_requests").Type != gjson.Null {
			requests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
		}
		costComponents = append(costComponents, stepFunctionExpressRequestCostComponent(region, requests))

		if u != nil && u.Get("workflow_duration_ms").Type != gjson.Null &&
			u.Get("monthly_requests").Type != gjson.Null &&
			u.Get("memory_mb").Type != gjson.Null {

			memoryRequest = decimalPtr(decimal.NewFromInt(u.Get("memory_mb").Int()))
			duration = decimalPtr(decimal.NewFromInt(u.Get("workflow_duration_ms").Int()))
			gbSeconds = decimalPtr(calculateStepFunctionGBSeconds(*memoryRequest, *duration, *requests))
			// first 1K GB-hours is 1000*60*60, next 4K GB-hours is 4000*60*60
			pushLimits := []int{3600000, 14400000}
			pushQuantities := usage.CalculateTierBuckets(*gbSeconds, pushLimits)

			costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (first 1K)", region, "0", &pushQuantities[0]))
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (next 4K)", region, "3600000", &pushQuantities[1]))
			}
			if pushQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (over 5K)", region, "18000000", &pushQuantities[2]))
			}
		} else {
			var unknown *decimal.Decimal
			costComponents = append(costComponents, stepFunctionExpressDurationCostComponent("Duration (first 1K)", region, "0", unknown))
		}
	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func stepFunctionStandardCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Transitions",
		Unit:            "1K transitions",
		UnitMultiplier:  decimal.NewFromInt(2),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonStates"),
			ProductFamily: strPtr("AWS Step Functions"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StateTransition/")},
			},
		},
	}
}

func stepFunctionExpressRequestCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonStates"),
			ProductFamily: strPtr("AWS Step Functions"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StepFunctions-Request/")},
			},
		},
	}
}

func stepFunctionExpressDurationCostComponent(name string, region string, startUsageAmt string, gbSeconds *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-hours",
		UnitMultiplier:  decimal.NewFromInt(3600),
		MonthlyQuantity: gbSeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonStates"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StepFunctions-GB-Second/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmt),
		},
	}
}

func calculateStepFunctionGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	// Use a min of 64MB, and round-up to nearest 64MB
	if memorySize.LessThan(decimal.NewFromInt(64)) {
		memorySize = decimal.NewFromInt(64)
	}
	roundedMemory := memorySize.Div(decimal.NewFromInt(64)).Ceil().Mul(decimal.NewFromInt(64))
	// Round up to nearest 100ms
	roundedDuration := averageRequestDuration.Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromInt(100))
	durationSeconds := monthlyRequests.Mul(roundedDuration).Mul(decimal.NewFromFloat(0.001))
	gbSeconds := durationSeconds.Mul(roundedMemory).Div(decimal.NewFromInt(1024))
	return gbSeconds
}
