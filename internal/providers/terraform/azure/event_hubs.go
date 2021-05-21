package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMEventHubsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_eventhub_namespace",
		RFunc: NewAzureRMEventHubs,
	}
}

func NewAzureRMEventHubs(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	sku := "Basic"
	location := d.Get("location").String()

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, eventHubsCostComponent("Ingress events", location, sku))
	costComponents = append(costComponents, eventHubsThroughPutCostComponent("Throughput unit (1 MB/s ingress, 2 MB/s egress)", location, sku))

	if sku == "Standard" {
		costComponents = append(costComponents, eventHubsKafkaCostComponent("Kafka Endpoint", location, sku))
		costComponents = append(costComponents, eventHubsCaptureCostComponent("Capture", location, sku))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eventHubsCostComponent(name, location, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", name, sku),
		Unit:            "hours",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Ingress Events", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsThroughPutCostComponent(name, location, sku string) *schema.CostComponent {

	return &schema.CostComponent{
		Name:           name,
		Unit:           "1M events",
		UnitMultiplier: 1,
		//MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Throughput Unit", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func eventHubsKafkaCostComponent(name, location, sku string) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "hours",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr("Kafka Endpoint")},
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
		Unit:            "hours",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Event Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Event Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr("Capture")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
