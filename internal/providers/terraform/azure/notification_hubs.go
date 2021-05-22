package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
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
	var monthlyAdditionalPushes *decimal.Decimal
	sku := "Basic"
	location := d.Get("location").String()

	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, notificationHubsCostComponent("Namespace usage", location, sku))
	if u != nil && u.Get("monthly_pushes").Type != gjson.Null {
		monthlyAdditionalPushes = decimalPtr(decimal.NewFromInt(u.Get("monthly_pushes").Int()))
		if monthlyAdditionalPushes.GreaterThanOrEqual(decimal.NewFromInt(10000000)) {
			if sku == "Basic" {
				costComponents = append(costComponents, notificationHubsPushesCostComponent("Additional pushes (Over 10M)", location, sku, "10", monthlyAdditionalPushes, 10000000))
			}
			if sku == "Standard" && monthlyAdditionalPushes.GreaterThan(decimal.NewFromInt(10000000)) {
				pushLimits := []int{10000000, 100000000}
				pushQuantities := usage.CalculateTierBuckets(*monthlyAdditionalPushes, pushLimits)
				if pushQuantities[1].GreaterThan(decimal.Zero) {
					newPushes := &pushQuantities[1]
					if pushQuantities[1].GreaterThanOrEqual(decimal.NewFromInt(100000000)) {
						remainingPushes := pushQuantities[1].Sub(pushQuantities[0])
						newPushes = &remainingPushes
					}
					costComponents = append(costComponents, notificationHubsPushesCostComponent("Additional pushes (10-100M)", location, sku, "10", newPushes, 1000000))
				}
				if pushQuantities[2].GreaterThan(decimal.Zero) {
					remainingPushes := pushQuantities[2].Add(pushQuantities[0])
					newPushes := &remainingPushes
					costComponents = append(costComponents, notificationHubsPushesCostComponent("Additional pushes (Over 100M)", location, sku, "100", newPushes, 1000000))
				}
			}
		}
	}
	if u == nil && sku != "Free" {
		costComponents = append(costComponents, notificationHubsPushesCostComponent("Additional pushes (10-100M)", location, sku, "10", monthlyAdditionalPushes, 1000000))

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func notificationHubsCostComponent(name, location, sku string) *schema.CostComponent {
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

func notificationHubsPushesCostComponent(name, location, sku, startUsageAmt string, quantity *decimal.Decimal, multi int) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(multi))))
	}
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1M pushes",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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
			StartUsageAmount: strPtr(startUsageAmt),
		},
	}
}
