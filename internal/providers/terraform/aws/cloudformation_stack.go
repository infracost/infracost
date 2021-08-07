package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetCloudFormationStackRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudformation_stack",
		RFunc: NewCloudFormationStack,
	}
}

func NewCloudFormationStack(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if d.Get("template_body").Type != gjson.Null && (checkAWS(d) || checkAlexa(d) || checkCustom(d)) {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cloudFormationCostComponents(d, u),
	}
}

func cloudFormationCostComponents(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	var monthlyHandlerOperations, monthlyDurationSecs *decimal.Decimal
	region := d.Get("region").String()

	if u != nil && u.Get("monthly_handler_operations").Type != gjson.Null {
		monthlyHandlerOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_handler_operations").Int()))
	}
	if u != nil && u.Get("monthly_duration_secs").Type != gjson.Null {
		monthlyDurationSecs = decimalPtr(decimal.NewFromInt(u.Get("monthly_duration_secs").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, cloudFormationCostComponent("Handler operations", region, "operations", "Resource-Invocation-Count", monthlyHandlerOperations))
	costComponents = append(costComponents, cloudFormationCostComponent("Durations above 30s", region, "seconds", "Resource-Processing-Time", monthlyDurationSecs))

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
func checkAWS(d *schema.ResourceData) bool {
	return strings.Contains(strings.ToLower(d.Get("template_body").String()), "aws::")
}
func checkAlexa(d *schema.ResourceData) bool {
	return strings.Contains(strings.ToLower(d.Get("template_body").String()), "alexa::")
}
func checkCustom(d *schema.ResourceData) bool {
	return strings.Contains(strings.ToLower(d.Get("template_body").String()), "custom::")
}
