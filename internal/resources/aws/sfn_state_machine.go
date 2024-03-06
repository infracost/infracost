package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"

	"strings"
)

type SFnStateMachine struct {
	Address            string
	Region             string
	Type               string
	MonthlyRequests    *int64 `infracost_usage:"monthly_requests"`
	WorkflowDurationMs *int64 `infracost_usage:"workflow_duration_ms"`
	MemoryMB           *int64 `infracost_usage:"memory_mb"`
	MonthlyTransitions *int64 `infracost_usage:"monthly_transitions"`
}

func (r *SFnStateMachine) CoreType() string {
	return "SFnStateMachine"
}

func (r *SFnStateMachine) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "workflow_duration_ms", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "memory_mb", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_transitions", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SFnStateMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SFnStateMachine) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	tier := r.Type
	if tier == "" {
		tier = "STANDARD"
	}

	if strings.ToLower(tier) == "standard" {
		var transitions *decimal.Decimal
		if r.MonthlyTransitions != nil {
			transitions = decimalPtr(decimal.NewFromInt(*r.MonthlyTransitions))
		}
		costComponents = append(costComponents, r.transistionsCostComponent(transitions))
	}

	if strings.ToLower(tier) == "express" {
		var requests *decimal.Decimal
		if r.MonthlyRequests != nil {
			requests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
		}
		costComponents = append(costComponents, r.requestsCostComponent(requests))

		if r.WorkflowDurationMs != nil && r.MonthlyRequests != nil && r.MemoryMB != nil {

			memoryRequest := decimalPtr(decimal.NewFromInt(*r.MemoryMB))
			duration := decimalPtr(decimal.NewFromInt(*r.WorkflowDurationMs))
			gbSeconds := decimalPtr(r.calculateGBSeconds(*memoryRequest, *duration, *requests))

			pushLimits := []int{3600000, 14400000}
			pushQuantities := usage.CalculateTierBuckets(*gbSeconds, pushLimits)

			costComponents = append(costComponents, r.durationCostComponent("Duration (first 1K)", "0", &pushQuantities[0]))
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.durationCostComponent("Duration (next 4K)", "3600000", &pushQuantities[1]))
			}
			if pushQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.durationCostComponent("Duration (over 5K)", "18000000", &pushQuantities[2]))
			}
		} else {
			var unknown *decimal.Decimal
			costComponents = append(costComponents, r.durationCostComponent("Duration (first 1K)", "0", unknown))
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *SFnStateMachine) transistionsCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Transitions",
		Unit:            "1K transitions",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonStates"),
			ProductFamily: strPtr("AWS Step Functions"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StateTransition/")},
			},
		},
		UsageBased: true,
	}
}

func (r *SFnStateMachine) requestsCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonStates"),
			ProductFamily: strPtr("AWS Step Functions"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StepFunctions-Request/")},
			},
		},
		UsageBased: true,
	}
}

func (r *SFnStateMachine) durationCostComponent(name string, startUsageAmt string, gbSeconds *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-hours",
		UnitMultiplier:  decimal.NewFromInt(3600),
		MonthlyQuantity: gbSeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonStates"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/StepFunctions-GB-Second/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmt),
		},
		UsageBased: true,
	}
}

func (r *SFnStateMachine) calculateGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {

	if memorySize.LessThan(decimal.NewFromInt(64)) {
		memorySize = decimal.NewFromInt(64)
	}
	roundedMemory := memorySize.Div(decimal.NewFromInt(64)).Ceil().Mul(decimal.NewFromInt(64))

	roundedDuration := averageRequestDuration.Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromInt(100))
	durationSeconds := monthlyRequests.Mul(roundedDuration).Mul(decimal.NewFromFloat(0.001))
	gbSeconds := durationSeconds.Mul(roundedMemory).Div(decimal.NewFromInt(1024))
	return gbSeconds
}
