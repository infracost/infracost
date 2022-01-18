package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type CloudformationStack struct {
	Address                  *string
	TemplateBody             *string
	Region                   *string
	MonthlyHandlerOperations *int64 `infracost_usage:"monthly_handler_operations"`
	MonthlyDurationSecs      *int64 `infracost_usage:"monthly_duration_secs"`
}

var CloudformationStackUsageSchema = []*schema.UsageItem{{Key: "monthly_handler_operations", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_duration_secs", ValueType: schema.Int64, DefaultValue: 0}}

func (r *CloudformationStack) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudformationStack) BuildResource() *schema.Resource {
	if r.TemplateBody != nil && (checkAWS(r.TemplateBody) || checkAlexa(r.TemplateBody) || checkCustom(r.TemplateBody)) {
		return &schema.Resource{
			Name:      *r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: CloudformationStackUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: cloudFormationCostComponents(r.Region, r.MonthlyHandlerOperations, r.MonthlyDurationSecs), UsageSchema: CloudformationStackUsageSchema,
	}
}

func cloudFormationCostComponents(region *string, rMonthlyHandlerOperations, rMonthlyDurationSecs *int64) []*schema.CostComponent {
	var monthlyHandlerOperations, monthlyDurationSecs *decimal.Decimal

	if rMonthlyHandlerOperations != nil {
		monthlyHandlerOperations = decimalPtr(decimal.NewFromInt(*rMonthlyHandlerOperations))
	}
	if rMonthlyDurationSecs != nil {
		monthlyDurationSecs = decimalPtr(decimal.NewFromInt(*rMonthlyDurationSecs))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, cloudFormationCostComponent("Handler operations", *region, "operations", "Resource-Invocation-Count", monthlyHandlerOperations))
	costComponents = append(costComponents, cloudFormationCostComponent("Durations above 30s", *region, "seconds", "Resource-Processing-Time", monthlyDurationSecs))

	return costComponents
}

func cloudFormationCostComponent(name, region, unit, usagetype string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AWSCloudFormation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", usagetype))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
func checkAWS(templateBody *string) bool {
	return strings.Contains(strings.ToLower(*templateBody), "aws::")
}
func checkAlexa(templateBody *string) bool {
	return strings.Contains(strings.ToLower(*templateBody), "alexa::")
}
func checkCustom(templateBody *string) bool {
	return strings.Contains(strings.ToLower(*templateBody), "custom::")
}
