package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_plan",
		RFunc: NewAzureRMAppServicePlan,
		Notes: []string{
			"Costs associated with running an app service plan in Azure",
		},
	}
}

func NewAzureRMAppServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: AppServicePlanCostComponent(d),
	}
}

// if d.Get("monitoring.0.enabled").Bool() {
// 	c := detailedMonitoringCostComponent(d)
// 	costComponents = append(costComponents, c)
// }

func AppServicePlanCostComponent(d *schema.ResourceData) []*schema.CostComponent {
	sku := d.Get("sku.0.size").String()
	purchaseOption := "Consumption"

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Computing usage (%s, %s)", purchaseOption, sku),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(d.Get("region").String()),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			Sku:           strPtr(sku),
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
			Unit:           strPtr("1 Hour"),
		},
	})

	return costComponents
}
