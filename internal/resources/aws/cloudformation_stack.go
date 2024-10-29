package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type CloudFormationStack struct {
	Address                  string
	Region                   string
	TemplateBody             string
	MonthlyHandlerOperations *int64 `infracost_usage:"monthly_handler_operations"`
	MonthlyDurationSecs      *int64 `infracost_usage:"monthly_duration_secs"`
}

func (r *CloudFormationStack) CoreType() string {
	return "CloudFormationStack"
}

func (r *CloudFormationStack) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_handler_operations", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_duration_secs", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *CloudFormationStack) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudFormationStack) BuildResource() *schema.Resource {
	if r.checkAWS() || r.checkAlexa() || r.checkCustom() {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: r.costComponents(),
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *CloudFormationStack) costComponents() []*schema.CostComponent {
	var monthlyHandlerOperations, monthlyDurationSecs *decimal.Decimal

	if r.MonthlyHandlerOperations != nil {
		monthlyHandlerOperations = decimalPtr(decimal.NewFromInt(*r.MonthlyHandlerOperations))
	}
	if r.MonthlyDurationSecs != nil {
		monthlyDurationSecs = decimalPtr(decimal.NewFromInt(*r.MonthlyDurationSecs))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, r.cloudFormationCostComponent("Handler operations", "operations", "Resource-Invocation-Count", monthlyHandlerOperations))
	costComponents = append(costComponents, r.cloudFormationCostComponent("Durations above 30s", "seconds", "Resource-Processing-Time", monthlyDurationSecs))

	return costComponents
}

func (r *CloudFormationStack) cloudFormationCostComponent(name, unit, usagetype string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AWSCloudFormation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", usagetype))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *CloudFormationStack) checkAWS() bool {
	return strings.Contains(strings.ToLower(r.TemplateBody), "aws::")
}

func (r *CloudFormationStack) checkAlexa() bool {
	return strings.Contains(strings.ToLower(r.TemplateBody), "alexa::")
}

func (r *CloudFormationStack) checkCustom() bool {
	return strings.Contains(strings.ToLower(r.TemplateBody), "custom::")
}
