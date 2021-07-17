package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMEventHubsNamespaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_eventhub_namespace",
		RFunc: NewAzureRMEventHubs,
		Notes: []string{"Premium namespaces are not supported by Terraform."},
	}
}

func NewAzureRMEventHubs(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var events, unknown *decimal.Decimal
	sku := "Basic"
	meterName := "Throughput Unit"
	region := lookupRegion(d, []string{})

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	capacity := decimal.NewFromInt(1)
	if u != nil && u.Get("throughput_or_capacity_units").Type != gjson.Null {
		capacity = decimal.NewFromInt(u.Get("throughput_or_capacity_units").Int())
	} else if d.Get("capacity").Type != gjson.Null {
		capacity = decimal.NewFromInt(d.Get("capacity").Int())
	}

	if d.Get("dedicated_cluster_id").Type != gjson.Null {
		sku = "Dedicated"
		meterName = "Capacity Unit"
	}
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_ingress_events").Type != gjson.Null {
		events = decimalPtr(decimal.NewFromInt(u.Get("monthly_ingress_events").Int()))
	}

	if strings.ToLower(sku) == "basic" {
		costComponents = append(costComponents, eventHubsIngressCostComponent(region, sku, events))
	}

	if strings.ToLower(sku) == "standard" {
		costComponents = append(costComponents, eventHubsIngressCostComponent(region, sku, events))
		if u != nil && u.Get("capture_enabled").Type != gjson.Null && strings.ToLower(u.Get("capture_enabled").String()) == "true" {
			costComponents = append(costComponents, eventHubsCaptureCostComponent(region, sku, capacity))
		}
	}

	costComponents = append(costComponents, eventHubsThroughPutCostComponent(region, sku, meterName, capacity))

	if strings.ToLower(sku) == "dedicated" {
		if u != nil && u.Get("retention_storage_gb").Type != gjson.Null {
			retention := decimalPtr(decimal.NewFromInt(u.Get("retention_storage_gb").Int()))
			// Subtract the 10 TB per capacity unit that's included in the dedicated namespace tier
			extendedRetention := retention.Sub(capacity.Mul(decimal.NewFromInt(10000)))
			if extendedRetention.GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent(region, sku, &extendedRetention))
			}
		} else {
			costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent(region, sku, unknown))
		}

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eventHubsIngressCostComponent(region, sku string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(1000000))))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Ingress event (%s)", sku),
		Unit:            "1M events",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s Ingress Events/i", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsThroughPutCostComponent(region, sku, meterName string, capacity decimal.Decimal) *schema.CostComponent {
	meterName = fmt.Sprintf("%s %s", sku, meterName)
	return &schema.CostComponent{
		Name:           "Throughput",
		Unit:           "units",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(capacity),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsCaptureCostComponent(region, sku string, quantity decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Capture",
		Unit:           "units",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", Value: strPtr("Capture")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsExtensionRetentionCostComponent(region, sku string, retention *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: retention,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Dedicated")},
				{Key: "meterName", Value: strPtr("Extended Retention")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
