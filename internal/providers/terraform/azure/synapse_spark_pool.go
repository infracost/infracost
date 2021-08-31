package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMSynapseSparkPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_synapse_spark_pool",
		RFunc: NewAzureRMSynapseSparkPool,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"the total costs consist of several resources that should be viewed as a whole"},
	}
}

func NewAzureRMSynapseSparkPool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	//region := lookupRegion(d, []string{"resource_group_name"})

	costComponents := make([]*schema.CostComponent, 0)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func synapseSparkPoolCostComponent(region, name, tier string, instances, vCores *decimal.Decimal) *schema.CostComponent {

	var HourlyQuantity *decimal.Decimal
	if instances != nil && vCores != nil {
		HourlyQuantity = decimalPtr(vCores.Mul(*instances))
	}

	return &schema.CostComponent{
		Name:           name,
		Unit:           "vCore",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: HourlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(fmt.Sprintf("Azure Synapse Analytics Data Flow - %s", tier))},
				{Key: "skuName", Value: strPtr("vCore")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
