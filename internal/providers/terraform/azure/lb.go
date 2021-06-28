package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetAzureRMLoadBalancerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb",
		RFunc: NewAzureRMLoadBalancer,
	}
}

func NewAzureRMLoadBalancer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := "Global"

	if strings.HasPrefix(strings.ToLower(d.Get("location").String()), "usgov") {
		region = "US Gov"
	} else if strings.Contains(strings.ToLower(d.Get("location").String()), "china") {
		region = "Сhina"
	}

	var costComponents []*schema.CostComponent
	var sku string
	var monthlyDataProcessedGb *decimal.Decimal

	if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_processed_gb").Int()))
	}

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	if strings.ToLower(sku) == "basic" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents = append(costComponents, dataProcessedCostComponent(region, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func dataProcessedCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Load Balancer"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Data Processed")},
			},
		},
	}
}
