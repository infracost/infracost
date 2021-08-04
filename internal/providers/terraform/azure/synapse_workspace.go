package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMSynapseWorkspacRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_synapse_workspace",
		RFunc: NewAzureRMSynapseWorkspace,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"the total costs consist of several resources that should be viewed as a whole"},
	}
}

func NewAzureRMSynapseWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})

	costComponents := make([]*schema.CostComponent, 0)

	var serverlessSqlPoolSize *decimal.Decimal
	if u != nil && u.Get("serverless_sql_pool_size_tb").Type != gjson.Null {
		serverlessSqlPoolSize = decimalPtr(decimal.NewFromInt(u.Get("serverless_sql_pool_size_tb").Int()))
	}

	if serverlessSqlPoolSize != nil && serverlessSqlPoolSize.LessThanOrEqual(decimal.NewFromInt(10)) {
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size (first 10TB)", "0", serverlessSqlPoolSize))
	}

	if serverlessSqlPoolSize != nil && serverlessSqlPoolSize.GreaterThan(decimal.NewFromInt(10)) {
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size (first 10TB)", "0", decimalPtr(decimal.NewFromInt(10))))
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size", "10", decimalPtr(serverlessSqlPoolSize.Sub(decimal.NewFromInt(10)))))
	}

	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - general purpose", "DZH318Z0CGNQ", decimal.NewFromInt(8), decimal.NewFromInt(1)))
	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - compute optimized", "DZH318Z0CGNF", decimal.NewFromInt(8), decimal.NewFromInt(1)))
	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - memory optimized", "DZH318Z0CGNM", decimal.NewFromInt(8), decimal.NewFromInt(1)))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func synapseServerlessSqlPoolCostComponent(region, name, start string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Synapse Analytics Serverless SQL Pool")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(start),
		},
	}
}

func synapseDataFlowCostComponent(region, name, productId string, vCores, instances decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "vCores",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(vCores.Mul(instances)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productId", Value: strPtr(productId)},
				{Key: "skuName", Value: strPtr("vCore")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

//db.products.distinct("attributes.description", {"vendorName": "azure", "service": "Azure Synapse Analytics", "productFamily": "Analytics" });
//db.products.find({"vendorName": "azure", "service": "Azure Synapse Analytics", "productFamily": "Analytics", "region": "westeurope", "attributes.productName": "Azure Synapse Analytics Serverless SQL Pool" }).pretty()
