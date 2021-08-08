package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMVpnGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_gateway",
		RFunc: NewAzureRMVpnGateway,
		Notes: []string{},
	}
}

func NewAzureRMVpnGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var tunnel, connection *decimal.Decimal
	sku := "Basic"
	// meterName := "Throughput Unit"
	region := lookupRegion(d, []string{})

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	costComponents := make([]*schema.CostComponent, 0)
	if sku == "Basic" {
		sku = "Basic Gateway"
	}
	if u != nil && u.Get("s2s_tunnel").Type != gjson.Null {
		tunnel = decimalPtr(decimal.NewFromInt(u.Get("s2s_tunnel").Int()))
	}
	if u != nil && u.Get("p2p_connection").Type != gjson.Null {
		connection = decimalPtr(decimal.NewFromInt(u.Get("p2p_connection").Int()))
	}
	if u != nil && u.Get("data_transfers").Type != gjson.Null {
		connection = decimalPtr(decimal.NewFromInt(u.Get("data_transfers").Int()))
	}
	costComponents = append(costComponents, vpnGateway(region, sku))

	// if tunnel != nil{
	// 	tunnelLimits := []int{10, 30}
	// }

	// if connection != nil {
	// 	connectionLimits := int{128, 250, 500, 1000, 5000, 10000}
	// }
	costComponents = append(costComponents, vpnGatewayS2S(region, sku, tunnel))
	costComponents = append(costComponents, vpnGatewayP2P(region, sku, connection))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGateway(region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Vpn Gateway (%s)", sku),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayS2S(region, sku string, tunnel *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:           "VPN Gateway S2S Tunnels",
		Unit:           "1 hour per tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		//HourlyQuantity: decimalPtr(capacity),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "S2S Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(""),
		},
	}
}

func vpnGatewayP2P(region, sku string, connection *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Gateway P2P Tunnels",
		Unit:           "1 hour per connection",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		//HourlyQuantity: decimalPtr(1),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "P2P Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(""),
		},
	}
}

func vpnGatewayDataTransfers(region, sku string, connection *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Gateway Data Tranfers",
		Unit:           "1 hour per connection",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		//HourlyQuantity: decimalPtr(1),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", sku))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "P2P Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(""),
		},
	}
}
