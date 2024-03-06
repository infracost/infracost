package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type APIManagement struct {
	Address                string
	Region                 string
	SKUName                string
	SelfHostedGatewayCount *int64 `infracost_usage:"self_hosted_gateway_count"`
	MonthlyAPICalls        *int64 `infracost_usage:"monthly_api_calls"`
}

func (r *APIManagement) CoreType() string {
	return "APIManagement"
}

func (r *APIManagement) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "self_hosted_gateway_count", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_api_calls", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *APIManagement) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIManagement) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}
	skuName := r.SKUName
	var tier string
	var capacity decimal.Decimal
	if s := strings.Split(skuName, "_"); len(s) == 2 {
		tier = strings.ToLower(s[0])
		capacity, _ = decimal.NewFromString(s[1])
	}

	if tier != "consumption" {
		costComponents = append(costComponents, r.apiManagementCostComponent(
			fmt.Sprintf("API management (%s)", tier),
			"units",
			tier,
			&capacity,
			false))

	} else {
		var apiCalls *decimal.Decimal
		if r.MonthlyAPICalls != nil {
			apiCalls = decimalPtr(decimal.NewFromInt(*r.MonthlyAPICalls))
		}

		if apiCalls != nil {
			apiCalls = decimalPtr(apiCalls.Div(decimal.NewFromInt(10000)))
			costComponents = append(costComponents, r.consumptionAPICostComponent(tier, apiCalls))
		} else {
			costComponents = append(costComponents, r.consumptionAPICostComponent(tier, apiCalls))
		}
	}

	if tier == "premium" {
		var selfHostedGateways *decimal.Decimal
		if r.SelfHostedGatewayCount != nil {
			selfHostedGateways = decimalPtr(decimal.NewFromInt(*r.SelfHostedGatewayCount))
		}
		costComponents = append(costComponents, r.apiManagementCostComponent(
			"Self hosted gateway",
			"gateways",
			"Gateway",
			selfHostedGateways,
			true,
		))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *APIManagement) apiManagementCostComponent(name, unit, tier string, quantity *decimal.Decimal, usageBased bool) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           unit,
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("API Management"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", tier))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s unit$/i", tier))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: usageBased,
	}
}

func (r *APIManagement) consumptionAPICostComponent(tier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "API management (consumption)",
		Unit:            "1M calls",
		UnitMultiplier:  decimal.NewFromInt(100),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("API Management"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", tier))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("100"),
		},
		UsageBased: true,
	}
}
