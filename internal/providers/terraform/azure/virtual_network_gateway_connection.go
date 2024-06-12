package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMVirtualNetworkGatewayConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_gateway_connection",
		RFunc: NewAzureRMVirtualNetworkGatewayConnection,
		ReferenceAttributes: []string{
			"virtual_network_gateway_id",
		},
		Notes: []string{"Price for additional S2S tunnels is used"},
	}
}

func NewAzureRMVirtualNetworkGatewayConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	sku := "Basic"

	var vpnGateway *schema.ResourceData
	if len(d.References("virtual_network_gateway_id")) > 0 {
		vpnGateway = d.References("virtual_network_gateway_id")[0]
		sku = vpnGateway.Get("sku").String()
	}

	region := d.Region
	if strings.ToLower(sku) == "basic" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}
	costComponents := make([]*schema.CostComponent, 0)

	if d.Get("type").Type != gjson.Null {
		if strings.ToLower(d.Get("type").String()) == "ipsec" {
			costComponents = append(costComponents, vpnGatewayS2S(region, sku))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGatewayS2S(region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("VPN gateway (%s)", sku),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
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
