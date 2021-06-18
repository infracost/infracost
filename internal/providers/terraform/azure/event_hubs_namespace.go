package azure

import (
	"fmt"

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
	var events, retention *decimal.Decimal
	var capacity int64
	sku := "Basic"
	meterName := "Throughput Unit"
	region := lookupRegion(d, []string{})

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}
	if d.Get("capacity").Type != gjson.Null {
		capacity = d.Get("capacity").Int()
	}
	if d.Get("dedicated_cluster_id").Type != gjson.Null {
		sku = "Dedicated"
		meterName = "Capacity Unit"
	}
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_ingress_events").Type != gjson.Null {
		events = decimalPtr(decimal.NewFromInt(u.Get("monthly_ingress_events").Int()))
	}

	costComponents = append(costComponents, eventHubsThroughPutCostComponent("Throughput", region, sku, meterName, capacity))

	if sku == "Basic" || sku == "Standard" {
		costComponents = append(costComponents, eventHubsCostComponent("Ingress events", region, sku, events))
	}
	if sku == "Standard" {
		costComponents = append(costComponents, eventHubsCaptureCostComponent("Capture", region, sku))
	}

	if sku == "Dedicated" {

		if u != nil && u.Get("extended_retention_storage_gb").Type != gjson.Null {
			retentionGB := []int{1000}
			retention = decimalPtr(decimal.NewFromInt(u.Get("extended_retention_storage_gb").Int()))
			retentionQuantites := usage.CalculateTierBuckets(*retention, retentionGB)
			if retentionQuantites[1].GreaterThanOrEqual(decimal.NewFromInt(1)) {
				costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent("Extended retention", region, sku, &retentionQuantites[1]))
			}
		} else {
			costComponents = append(costComponents, eventHubsExtensionRetentionCostComponent("Extended retention", region, sku, retention))
		}

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eventHubsCostComponent(name, location, sku string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(1000000))))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", name, sku),
		Unit:            "1M events",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("%s/i Ingress Events", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsThroughPutCostComponent(name, location, sku, meterName string, capacity int64) *schema.CostComponent {
	meterName = fmt.Sprintf("%s %s", sku, meterName)
	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s (%v)", name, capacity),
		Unit:           "units",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		//MonthlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsCaptureCostComponent(name, location, sku string) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "hour",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", Value: strPtr("Capture")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsExtensionRetentionCostComponent(name, location, sku string, retentionQuantites *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: retentionQuantites,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("%s/i", sku))},
				{Key: "meterName", Value: strPtr("Extended Retention")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
