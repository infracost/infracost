package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMVpnGatewayConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_gateway_connection",
		RFunc: NewAzureRMVpnGatewayConnection,
		ReferenceAttributes: []string{
			"type",
		},
		Notes: []string{"Price for additional S2S tunnels is used"},
	}
}

func NewAzureRMVpnGatewayConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	sku := "Basic"

	var vpnGateway *schema.ResourceData
	if len(d.References("type")) > 0 {
		vpnGateway = d.References("type")[0]
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
		if strings.ToLower(d.Get("type").String()) == "ipsec" {
			costComponents = append(costComponents, vpnGatewayS2S(region, sku, meterName))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGatewayS2S(region, sku, meterName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN gateway S2S tunnel",
		Unit:           "tunnel",
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
