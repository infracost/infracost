package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMPublicIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_public_ip",
		RFunc: NewAzureRMPublicIP,
	}
}

func NewAzureRMPublicIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var meterName string
	location := d.Get("location").String()
	sku := "Basic"
	allocationMethod := d.Get("allocation_method").String()

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	switch sku {
	case "Basic":
		meterName = allocationMethod + " Public IP"
	case "Standard":
		meterName = "Standard " + allocationMethod + " Public IP"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, PublicIPCostComponent(fmt.Sprintf("IP address (%s)", strings.ToLower(allocationMethod)), location, sku, meterName))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func PublicIPCostComponent(name, location, sku, meterName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("IP Addresses")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
