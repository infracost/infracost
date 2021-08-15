package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMVpnGatewayConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_gateway_connection",
		RFunc: NewAzureRMVpnGatewayConnection,
		Notes: []string{},
	}
}

func NewAzureRMVpnGatewayConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	fmt.Println("============================================")
	var tunnel *decimal.Decimal
	sku := "Basic"
	region := lookupRegion(d, []string{})
	fmt.Println(u)

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}
	meterName := sku

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "Basic" {
		sku = "Basic Gateway"
		meterName = "Basic"
	}

	if u != nil && u.Get("s2s_tunnel").Type != gjson.Null {
		tunnel = decimalPtr(decimal.NewFromInt(u.Get("s2s_tunnel").Int()))
		if tunnel != nil {
			tunnelLimits := []int{10}
			tunnelValues := usage.CalculateTierBuckets(*tunnel, tunnelLimits)
			if tunnelValues[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayS2S(region, sku, meterName, &tunnelValues[1]))
			}
		}
	} else {
		costComponents = append(costComponents, vpnGatewayS2S(region, sku, meterName, tunnel))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGatewayS2S(region, sku, meterName string, tunnel *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN gateway S2S tunnels (over 10)",
		Unit:           "tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: tunnel,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "S2S Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
