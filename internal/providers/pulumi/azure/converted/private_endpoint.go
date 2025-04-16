package azure

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMPrivateEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_endpoint",
		RFunc: NewAzureRMPrivateEndpoint,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"Is connected to the free item private link service."},
	}
}

func NewAzureRMPrivateEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region
	region = convertRegion(region)

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, privateEndpointCostComponent(region, "Private endpoint", "Standard Private Endpoint"))

	if u != nil && u.Get("monthly_inbound_data_processed_gb").Type != gjson.Null {
		inbound := decimal.NewFromInt(u.Get("monthly_inbound_data_processed_gb").Int())

		inboundTiers := []int{1_000_000, 4_000_000}
		inboundQuantities := usage.CalculateTierBuckets(inbound, inboundTiers)

		if len(inboundQuantities) > 0 {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Inbound data processed (first 1PB)", "Standard Data Processed - Ingress", "0", &inboundQuantities[0]))
		}

		if len(inboundQuantities) > 1 && inboundQuantities[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Inbound data processed (next 4PB)", "Standard Data Processed - Ingress", "1000000", &inboundQuantities[1]))
		}

		if len(inboundQuantities) > 2 && inboundQuantities[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Inbound data processed (over 5PB)", "Standard Data Processed - Ingress", "5000000", &inboundQuantities[2]))
		}
	} else {
		costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Inbound data processed (first 1PB)", "Standard Data Processed - Ingress", "0", nil))
	}

	if u != nil && u.Get("monthly_outbound_data_processed_gb").Type != gjson.Null {
		outbound := decimal.NewFromInt(u.Get("monthly_outbound_data_processed_gb").Int())

		outboundTiers := []int{1_000_000, 4_000_000}
		outboundQuantities := usage.CalculateTierBuckets(outbound, outboundTiers)

		if len(outboundQuantities) > 0 {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Outbound data processed (first 1PB)", "Standard Data Processed - Egress", "0", &outboundQuantities[0]))
		}

		if len(outboundQuantities) > 1 && outboundQuantities[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Outbound data processed (next 4PB)", "Standard Data Processed - Egress", "1000000", &outboundQuantities[1]))
		}

		if len(outboundQuantities) > 2 && outboundQuantities[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Outbound data processed (over 5PB)", "Standard Data Processed - Egress", "5000000", &outboundQuantities[2]))
		}
	} else {
		costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Outbound data processed (first 1PB)", "Standard Data Processed - Egress", "0", nil))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func privateEndpointCostComponent(region, name, meterName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Virtual Network Private Link")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
	}
}

func privateEndpointDataCostComponent(region, name, meterName, startUsage string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Virtual Network Private Link")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
