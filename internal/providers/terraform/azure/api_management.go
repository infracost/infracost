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
	costComponents := []*schema.CostComponent{}
	location := d.Get("location").String()
	skuName := d.Get("sku_name").String()
	var tier string
	var capacity decimal.Decimal
	if s := strings.Split(skuName, "_"); len(s) == 2 {
		tier = s[0]
		capacity, _ = decimal.NewFromString(s[1])
	}

	if tier != "Consumption" {
		costComponents = append(costComponents, apiManagementCostComponent(
			fmt.Sprintf("API management (%s)", tier),
			"units",
			location,
			tier,
			&capacity))

	} else {
		var apiCalls *decimal.Decimal
		if u != nil && u.Get("monthly_api_calls").Type != gjson.Null {
			apiCalls = decimalPtr(decimal.NewFromInt(u.Get("monthly_api_calls").Int()))
		}

		if apiCalls != nil {
			if apiCalls.GreaterThan(decimal.NewFromInt(1_000_000)) {
				apiCalls = decimalPtr(apiCalls.Sub(decimal.NewFromInt(1_000_000)).Div(decimal.NewFromInt(10000)))
				costComponents = append(costComponents, consumptionApiCostComponent(location, tier, apiCalls))
			}
		} else {
			costComponents = append(costComponents, consumptionApiCostComponent(location, tier, apiCalls))
		}
	}

	if tier == "Premium" {
		var selfHostedGateways *decimal.Decimal
		if u != nil && u.Get("self_hosted_gateway_count").Type != gjson.Null {
			selfHostedGateways = decimalPtr(decimal.NewFromInt(u.Get("self_hosted_gateway_count").Int()))
		}
		costComponents = append(costComponents, apiManagementCostComponent(
			"Self hosted gateway",
			"gateways",
			location,
			"Gateway",
			selfHostedGateways,
		))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func apiManagementCostComponent(name, unit, location, tier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           unit,
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("API Management"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(tier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func consumptionApiCostComponent(location, tier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "API management (consumption over 1M)",
		Unit:            "1M calls",
		UnitMultiplier:  100,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("API Management"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(tier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("100"),
		},
	}
}
