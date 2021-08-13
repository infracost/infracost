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
	var tunnel, connection, dataTransfers *decimal.Decimal
	sku := "Basic"
	region := lookupRegion(d, []string{})
	zone := regionToZone(region)

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}
	meterName := sku

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "Basic" {
		sku = "Basic Gateway"
		meterName = "Basic"
	}

	costComponents = append(costComponents, vpnGateway(region, sku, meterName))

	if u != nil && u.Get("s2s_tunnel").Type != gjson.Null {
		tunnel = decimalPtr(decimal.NewFromInt(u.Get("s2s_tunnel").Int()))
		if tunnel != nil {
			tunnelLimits := []int{10}
			tunnelValues := usage.CalculateTierBuckets(*tunnel, tunnelLimits)
			if tunnelValues[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayS2S(region, sku, &tunnelValues[1]))
			}
		}
	} else {
		costComponents = append(costComponents, vpnGatewayS2S(region, sku, tunnel))
	}

	if u != nil && u.Get("p2s_connection").Type != gjson.Null {
		connection = decimalPtr(decimal.NewFromInt(u.Get("p2s_connection").Int()))
		if connection != nil {
			connectionLimits := []int{128}
			connectionValues := usage.CalculateTierBuckets(*connection, connectionLimits)
			fmt.Println("connectionValues", connectionValues)
			if connectionValues[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, vpnGatewayP2S(region, sku, &connectionValues[1]))
			}
		}
	} else {
		costComponents = append(costComponents, vpnGatewayP2S(region, sku, connection))
	}

	if u != nil && u.Get("monthly_data_transfer_gb").Type != gjson.Null {
		dataTransfers = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_transfer_gb").Int()))
		if dataTransfers != nil {
			costComponents = append(costComponents, vpnGatewayDataTransfers(zone, sku, dataTransfers))
		}
	} else {
		costComponents = append(costComponents, vpnGatewayDataTransfers(zone, sku, dataTransfers))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpnGateway(region, sku, meterName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("VPN gateway (%s)", meterName),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(sku)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayS2S(region, sku string, tunnel *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:           "VPN gateway S2S tunnels",
		Unit:           "tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: tunnel,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(sku)},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "S2S Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayP2S(region, sku string, connection *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN gateway P2S tunnels",
		Unit:           "tunnel",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: connection,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(region),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(sku)},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "P2S Connection"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func vpnGatewayDataTransfers(zone, sku string, dataTransfers *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "VPN gateway data tranfer",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataTransfers,
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
