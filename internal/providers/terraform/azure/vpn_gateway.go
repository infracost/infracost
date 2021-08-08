package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
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
	var tunnel, connection, data_transfers *decimal.Decimal
	sku := "Basic"
	region := lookupRegion(d, []string{})
	zone := regionToZone(region)

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "Basic" {
		sku = "Basic Gateway"
	}

	costComponents = append(costComponents, vpnGateway(region, sku))

	if u != nil && u.Get("s2s_tunnel").Type != gjson.Null {
		tunnel = decimalPtr(decimal.NewFromInt(u.Get("s2s_tunnel").Int()))
		if tunnel != nil {
			tunnelLimits := []int{10, 30}
			tunnelValues := usage.CalculateTierBuckets(*tunnel, tunnelLimits)
			if tunnelValues[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayS2S(region, sku, &tunnelValues[1]))
			}
		}
	} else {
		costComponents = append(costComponents, vpnGatewayS2S(region, sku, tunnel))
	}

	if u != nil && u.Get("p2p_connection").Type != gjson.Null {
		connection = decimalPtr(decimal.NewFromInt(u.Get("p2p_connection").Int()))
		if connection != nil {
			connectionLimits := []int{128, 250, 500, 1000, 5000, 10000}
			connectionValues := usage.CalculateTierBuckets(*connection, connectionLimits)
			if sku == "VpnGw1" && connectionValues[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2P(region, sku, &connectionValues[1]))
			} else if sku == "VpnGw2" && connectionValues[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2P(region, sku, &connectionValues[2]))
			} else if sku == "VpnGw3" && connectionValues[3].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2P(region, sku, &connectionValues[3]))
			} else if sku == "VpnGw4" && connectionValues[4].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2P(region, sku, &connectionValues[4]))
			} else if sku == "VpnGw5" && connectionValues[5].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2P(region, sku, &connectionValues[5]))
			}
		}
	} else {
		costComponents = append(costComponents, vpnGatewayP2P(region, sku, connection))
	}

	if u != nil && u.Get("data_transfers").Type != gjson.Null {
		data_transfers = decimalPtr(decimal.NewFromInt(u.Get("data_transfers").Int()))
		if data_transfers != nil {
			costComponents = append(costComponents, vpnGatewayDataTransfers(zone, sku, data_transfers))
		}
	} else {
		costComponents = append(costComponents, vpnGatewayDataTransfers(zone, sku, data_transfers))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGateway(region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("VPN Gateway (%s)", sku),
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
		Unit:           "tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: tunnel,
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
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayP2P(region, sku string, connection *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Gateway P2P Tunnels",
		Unit:           "tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: connection,
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
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayDataTransfers(zone, sku string, data_transfers *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Gateway Data Tranfers",
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: data_transfers,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(zone),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "serviceFamily", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "Networking"))},
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "VPN Gateway Bandwidth"))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "Inter-Virtual Network Data Transfer Out"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
