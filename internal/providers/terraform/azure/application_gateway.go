package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMApplicationGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_application_gateway",
		RFunc: NewAzureRMApplicationGateway,
	}
}

func NewAzureRMApplicationGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	var monthlyDataProcessedGb *decimal.Decimal
	skuName := d.Get("sku.0.name").String()
	var sku, tier string
	costComponents := make([]*schema.CostComponent, 0)
	tierLimits := []int{10240, 30720}

	var capacityUnits int64 = 1

	if d.Get("autoscale_configuration.0.min_capacity").Type != gjson.Null {
		capacityUnits = d.Get("autoscale_configuration.0.min_capacity").Int()
	}

	if u != nil && u.Get("capacity_units").Type != gjson.Null {
		capacityUnits = u.Get("capacity_units").Int()
	}

	if d.Get("sku.0.capacity").Type != gjson.Null {
		capacityUnits = d.Get("sku.0.capacity").Int()
	}

	skuNameParts := strings.Split(skuName, "_")
	if len(skuNameParts) > 1 {
		sku = strings.ToLower(skuNameParts[1])
	}

	if sku != "v2" {
		if strings.ToLower(skuNameParts[0]) == "standard" {
			tier = "basic"
		} else {
			tier = "WAF"
		}
		costComponents = append(costComponents, gatewayCostComponent(fmt.Sprintf("Gateway usage (%s, %s)", tier, sku), region, tier, sku, capacityUnits))

		if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
			monthlyDataProcessedGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_processed_gb").Int()))
			result := usage.CalculateTierBuckets(*monthlyDataProcessedGb, tierLimits)

			if sku == "small" {
				if result[0].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (0-10TB)", region, sku, "0", &result[0]))
				}
				if result[1].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (10-40TB)", region, sku, "0", &result[1]))
				}
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (over 40TB)", region, sku, "0", &result[2]))
				}
			}

			if sku == "medium" {
				if result[1].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (10-40TB)", region, sku, "10240", &result[1]))
				}
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (over 40TB)", region, sku, "10240", &result[2]))
				}
			}

			if sku == "large" {
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, dataProcessingCostComponent("Data processing (over 40TB)", region, sku, "40960", &result[2]))
				}
			}

		} else {
			var unknown *decimal.Decimal
			costComponents = append(costComponents, dataProcessingCostComponent("Data processing (0-10TB)", region, sku, "0", unknown))
		}
	}

	if sku == "v2" {
		if strings.ToLower(skuNameParts[0]) == "standard" {
			tier = "basic v2"
			costComponents = append(costComponents, fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), region, "standard v2"))
			costComponents = append(costComponents, capacityUnitsCostComponent("basic", region, "standard v2", capacityUnits))
		} else {
			tier = "WAF v2"
			costComponents = append(costComponents, fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), region, tier))
			costComponents = append(costComponents, capacityUnitsCostComponent("WAF", region, tier, capacityUnits))
		}

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func gatewayCostComponent(name, region, tier, sku string, capacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacityUnits)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("%s Application Gateway$", tier))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Gateway$", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func dataProcessingCostComponent(name, region, sku, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Data Processed", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
func capacityUnitsCostComponent(name, region, tier string, capacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("V2 capacity units (%s)", name),
		Unit:           "CU",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacityUnits)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("Application Gateway %s$", tier))},
				{Key: "meterName", ValueRegex: regexPtr("Capacity Units$")},
			},
		},

		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func fixedForV2CostComponent(name, region, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("Application Gateway %s$", tier))},
				{Key: "meterName", ValueRegex: regexPtr("Fixed Cost$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
