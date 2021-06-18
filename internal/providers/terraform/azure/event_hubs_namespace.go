package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMEventHubsNamespaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_eventhub_namespace",
		RFunc: NewAzureRMEventHubs,
	}
}

func NewAzureRMEventHubs(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var events, retention, capacityUnit *decimal.Decimal
	var capacity int64
	sku := "Basic"
	meterName := "Throughput Unit"
	region := lookupRegion(d, []string{})

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}
	if d.Get("capacity").Type != gjson.Null {
		capacity = d.Get("capacity").Int()
		capacityUnit = decimalPtr(decimal.NewFromInt(capacity))
	}
	if u != nil && u.Get("throughput_or_capacity_units").Type != gjson.Null {
		capacityUnit = decimalPtr(decimal.NewFromInt(u.Get("throughput_or_capacity_units").Int()))
	}

	if d.Get("dedicated_cluster_id").Type != gjson.Null {
		sku = "Dedicated"
		meterName = "Capacity Unit"
	}
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_ingress_events").Type != gjson.Null {
		events = decimalPtr(decimal.NewFromInt(u.Get("monthly_ingress_events").Int()))
	}

	costComponents = append(costComponents, eventHubsThroughPutCostComponent(region, sku, meterName, capacityUnit))

	if strings.ToLower(sku) == "basic" {
		costComponents = append(costComponents, eventHubsCostComponent(region, sku, events))
	}
	if strings.ToLower(sku) == "standard" {
		costComponents = append(costComponents, eventHubsCostComponent(region, sku, events))
		costComponents = append(costComponents, eventHubsCaptureCostComponent(region, sku))
	}

	if strings.ToLower(sku) == "dedicated" {

		if u != nil && u.Get("extended_retention_storage_gb").Type != gjson.Null {
			retentionGB := []int{1000}
			retention = decimalPtr(decimal.NewFromInt(u.Get("extended_retention_storage_gb").Int()))
			retentionQuantites := usage.CalculateTierBuckets(*retention, retentionGB)
			if retentionQuantites[1].GreaterThanOrEqual(decimal.NewFromInt(1)) {
				costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent(region, sku, &retentionQuantites[1]))
			}
		} else {
			costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent(region, sku, retention))
		}

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eventHubsCostComponent(region, sku string, quantity *decimal.Decimal) *schema.CostComponent {
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

func eventHubsThroughPutCostComponent(region, sku, meterName string, capacityUnit *decimal.Decimal) *schema.CostComponent {
	meterName = fmt.Sprintf("%s %s", sku, meterName)
	return &schema.CostComponent{
		Name:            "Throughput",
		Unit:            "units",
		UnitMultiplier:  1,
		MonthlyQuantity: capacityUnit,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsCaptureCostComponent(region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Capture",
		Unit:           "hour",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", Value: strPtr("Capture")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsExtensionRetentionCostComponent(region, sku string, retentionQuantites *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: retentionQuantites,
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
