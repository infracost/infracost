package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMApiManagementRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_api_management",
		RFunc: NewAzureRMApiManagement,
		ReferenceAttributes: []string{
			"certificate_id",
		},
	}
}

func NewAzureRMApiManagement(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	costComponents := []*schema.CostComponent{}
	skuName := d.Get("sku_name").String()
	var tier string
	var capacity decimal.Decimal
	if s := strings.Split(skuName, "_"); len(s) == 2 {
		tier = strings.ToLower(s[0])
		capacity, _ = decimal.NewFromString(s[1])
	}

	if tier != "consumption" {
		costComponents = append(costComponents, apiManagementCostComponent(
			fmt.Sprintf("API management (%s)", tier),
			"units",
			region,
			tier,
			&capacity))

	} else {
		var apiCalls *decimal.Decimal
		if u != nil && u.Get("monthly_api_calls").Type != gjson.Null {
			apiCalls = decimalPtr(decimal.NewFromInt(u.Get("monthly_api_calls").Int()))
		}

		if apiCalls != nil {
			apiCalls = decimalPtr(apiCalls.Div(decimal.NewFromInt(10000)))
			costComponents = append(costComponents, consumptionAPICostComponent(region, tier, apiCalls))
		} else {
			costComponents = append(costComponents, consumptionAPICostComponent(region, tier, apiCalls))
		}
	}

	if tier == "premium" {
		var selfHostedGateways *decimal.Decimal
		if u != nil && u.Get("self_hosted_gateway_count").Type != gjson.Null {
			selfHostedGateways = decimalPtr(decimal.NewFromInt(u.Get("self_hosted_gateway_count").Int()))
		}
		costComponents = append(costComponents, apiManagementCostComponent(
			"Self hosted gateway",
			"gateways",
			region,
			"Gateway",
			selfHostedGateways,
		))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func apiManagementCostComponent(name, unit, region, tier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           unit,
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("API Management"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", tier))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func consumptionAPICostComponent(region, tier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "API management (consumption)",
		Unit:            "1M calls",
		UnitMultiplier:  decimal.NewFromInt(100),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
	}
}
