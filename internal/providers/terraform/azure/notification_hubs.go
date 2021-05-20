package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMNotificationHubsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_notification_hub_namespace",
		RFunc: NewAzureRMNotificationHubs,
	}
}

func NewAzureRMNotificationHubs(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	sku := "Basic"
	location := d.Get("location").String()

	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, NotificationHubsCostComponent("Base Charge Per Namespace", location, sku))
	costComponents = append(costComponents, NotificationHubsPushesCostComponent("Additional Pushes (over 10M)", location, sku))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func NotificationHubsCostComponent(name, location, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", name, sku),
		Unit:            "months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Notification Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Notification Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Unit", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func NotificationHubsPushesCostComponent(name, location, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "Million Pushes",
		UnitMultiplier: 1,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(location),
			Service:    strPtr("Notification Hubs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Notification Hubs")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Pushes", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("10"),
		},
	}
}
