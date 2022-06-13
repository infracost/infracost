package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMBastionHostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_bastion_host",
		RFunc: NewAzureRMBastionHost,
	}
}

func NewAzureRMBastionHost(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	productType := "Basic"
	region := lookupRegion(d, []string{})

	costComponents := []*schema.CostComponent{
		{
			Name:           "Bastion host",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Azure Bastion"),
				ProductFamily: strPtr("Networking"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr(productType)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Gateway", productType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}

	costComponents = append(costComponents, outboundDataTransferComponents(u, region, productType)...)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func outboundDataTransferComponents(u *schema.UsageData, region, productType string) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)
	name := "Outbound data transfer "

	type dataTier struct {
		name       string
		startUsage string
	}

	data := []dataTier{
		{name: fmt.Sprintf("%s%s", name, "(first 10TB)"), startUsage: "5"},
		{name: fmt.Sprintf("%s%s", name, "(next 40TB)"), startUsage: "10240"},
		{name: fmt.Sprintf("%s%s", name, "(next 100TB)"), startUsage: "51200"},
		{name: fmt.Sprintf("%s%s", name, "(next 350TB)"), startUsage: "153600"},
		{name: fmt.Sprintf("%s%s", name, "(over 500TB)"), startUsage: "512000"},
	}

	if u != nil && u.Get("monthly_outbound_data_gb").Exists() {
		tierLimits := []int{10000, 50000, 150000, 500000}
		monthlyOutboundDataGb := decimal.NewFromInt(u.Get("monthly_outbound_data_gb").Int())
		tiers := usage.CalculateTierBuckets(monthlyOutboundDataGb, tierLimits)
		for i, d := range data {
			if i < len(tiers) && tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, outboundDataTransferSingleComponent(
					d.name,
					region,
					productType,
					d.startUsage,
					decimalPtr(tiers[i])))
			}
		}
	} else {
		costComponents = append(costComponents, outboundDataTransferSingleComponent(
			data[0].name,
			region,
			productType,
			data[0].startUsage,
			nil))
	}
	return costComponents
}

func outboundDataTransferSingleComponent(name, region, productType, startUsage string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Bastion"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(productType)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Data Transfer Out", productType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
