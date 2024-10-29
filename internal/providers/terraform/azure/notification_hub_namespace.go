package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMNotificationHubNamespaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_notification_hub_namespace",
		RFunc: NewAzureRMNotificationHubNamespace,
	}
}

func NewAzureRMNotificationHubNamespace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var monthlyAdditionalPushes *decimal.Decimal
	sku := "Basic"

	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}
	if strings.ToLower(sku) == "free" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, notificationHubsCostComponent("Namespace usage", region, sku))
	if u != nil && u.Get("monthly_pushes").Type != gjson.Null {
		monthlyAdditionalPushes = decimalPtr(decimal.NewFromInt(u.Get("monthly_pushes").Int()))
	}

	if strings.ToLower(sku) == "basic" {
		if monthlyAdditionalPushes != nil {
			pushLimits := []int{10000000}
			pushQuantities := usage.CalculateTierBuckets(*monthlyAdditionalPushes, pushLimits)
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (over 10M)", region, sku, "10", &pushQuantities[1], 1000000))
			}
		} else {
			costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (over 10M)", region, sku, "10", nil, 1000000))
		}
	} else {
		if monthlyAdditionalPushes != nil {
			pushLimits := []int{10000000, 90000000}
			pushQuantities := usage.CalculateTierBuckets(*monthlyAdditionalPushes, pushLimits)
			if pushQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (10-100M)", region, sku, "10", &pushQuantities[1], 1000000))
			}
			if pushQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (over 100M)", region, sku, "100", &pushQuantities[2], 1000000))
			}
		} else {
			costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (10-100M)", region, sku, "10", nil, 1000000))
			costComponents = append(costComponents, notificationHubsPushesCostComponent("Pushes (over 100M)", region, sku, "100", nil, 1000000))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func notificationHubsCostComponent(name, region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", name, sku),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
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

func notificationHubsPushesCostComponent(name, region, sku, startUsageAmt string, quantity *decimal.Decimal, multi int) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(multi))))
	}
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1M pushes",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
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
