package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
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
	region := lookupRegion(d, []string{"resource_group_name"})
	region = convertRegion(region)

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, privateEndpointCostComponent(region, "Private endpoint", "Private Endpoint"))

	var inbound, outbound *decimal.Decimal
	if u != nil && u.Get("monthly_inbound_data_processed_gb").Type != gjson.Null {
		inbound = decimalPtr(decimal.NewFromInt(u.Get("monthly_inbound_data_processed_gb").Int()))
	}
	costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Inbound data processed", "Data Processed - Ingress", inbound))

	if u != nil && u.Get("monthly_outbound_data_processed_gb").Type != gjson.Null {
		outbound = decimalPtr(decimal.NewFromInt(u.Get("monthly_outbound_data_processed_gb").Int()))
	}
	costComponents = append(costComponents, privateEndpointDataCostComponent(region, "Outbound data processed", "Data Processed - Egress", outbound))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func privateEndpointCostComponent(region, name, meterName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 name,
		Unit:                 "hour",
		UnitMultiplier:       1,
		MonthlyQuantity:      decimalPtr(decimal.NewFromInt(730)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Virtual Network Private Link")},
				{Key: "meterName", ValueRegex: strPtr(meterName)},
			},
		},
	}
}

func privateEndpointDataCostComponent(region, name, meterName string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 name,
		Unit:                 "GB",
		UnitMultiplier:       1,
		MonthlyQuantity:      quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Virtual Network Private Link")},
				{Key: "meterName", ValueRegex: strPtr(meterName)},
			},
		},
	}
}
