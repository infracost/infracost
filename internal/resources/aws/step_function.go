package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"

	"strings"
)

type SfnStateMachine struct {
	Address            *string
	Region             *string
	Type               *string
	MonthlyRequests    *int64 `infracost_usage:"monthly_requests"`
	WorkflowDurationMs *int64 `infracost_usage:"workflow_duration_ms"`
	MemoryMb           *int64 `infracost_usage:"memory_mb"`
	MonthlyTransitions *int64 `infracost_usage:"monthly_transitions"`
}

var SfnStateMachineUsageSchema = []*schema.UsageItem{{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "workflow_duration_ms", ValueType: schema.Int64, DefaultValue: 0}, {Key: "memory_mb", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_transitions", ValueType: schema.Int64, DefaultValue: 0}}

func (r *SfnStateMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SfnStateMachine) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)

	var duration, memoryRequest, requests, transitions, gbSeconds *decimal.Decimal

	tier := "STANDARD"
	if r.Type != nil {
		tier = *r.Type
	}

	if strings.ToLower(tier) == "standard" {
		if r.MonthlyTransitions != nil {
			transitions = decimalPtr(decimal.NewFromInt(*r.MonthlyTransitions))
		}
		costComponents = append(costComponents, stepFunctionStandardCostComponent(region, transitions))
	}

	if strings.ToLower(tier) == "express" {
		if r.MonthlyRequests != nil {
			requests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
		}
		costComponents = append(costComponents, stepFunctionExpressRequestCostComponent(region, requests))

		if r != nil && r.WorkflowDurationMs != nil && r.MonthlyRequests != nil && r.MemoryMb != nil {

			memoryRequest = decimalPtr(decimal.NewFromInt(*r.MemoryMb))
			duration = decimalPtr(decimal.NewFromInt(*r.WorkflowDurationMs))
			gbSeconds = decimalPtr(calculateStepFunctionGBSeconds(*memoryRequest, *duration, *requests))

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
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: SfnStateMachineUsageSchema,
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

	if memorySize.LessThan(decimal.NewFromInt(64)) {
		memorySize = decimal.NewFromInt(64)
	}
	roundedMemory := memorySize.Div(decimal.NewFromInt(64)).Ceil().Mul(decimal.NewFromInt(64))

	roundedDuration := averageRequestDuration.Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromInt(100))
	durationSeconds := monthlyRequests.Mul(roundedDuration).Mul(decimal.NewFromFloat(0.001))
	gbSeconds := durationSeconds.Mul(roundedMemory).Div(decimal.NewFromInt(1024))
	return gbSeconds
}
