package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMVpnGatewayConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_gateway_connection",
		RFunc: NewAzureRMVpnGatewayConnection,
		ReferenceAttributes: []string{
			"virtual_network_gateway_id",
		},
		Notes: []string{"Price for additional S2S tunnels is used"},
	}
}

func NewAzureRMVpnGatewayConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	var tunnel *decimal.Decimal
	sku := "Basic"

	var vpnGateway *schema.ResourceData
	if len(d.References("virtual_network_gateway_id")) > 0 {
		vpnGateway = d.References("virtual_network_gateway_id")[0]
		sku = vpnGateway.Get("sku").String()
	}

	region := lookupRegion(d, []string{})

	meterName := sku

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "Basic" {
		sku = "Basic Gateway"
		meterName = "Basic"
	}

	if d.Get("type").Type != gjson.Null {
		if strings.ToLower(d.Get("type").String()) == "ipsec" && u != nil && u.Get("s2s_tunnel").Type != gjson.Null {
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
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGatewayS2S(region, sku, meterName string, tunnel *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN gateway S2S tunnel",
		Unit:           "tunnel",
		UnitMultiplier: decimal.NewFromInt(1),
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
