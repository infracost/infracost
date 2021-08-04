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

	managedVirtualNetwork := false
	if d.Get("managed_virtual_network_enabled").Type != gjson.Null {
		managedVirtualNetwork = d.Get("managed_virtual_network_enabled").Bool()
	}

	if managedVirtualNetwork {
		costComponents = append(costComponents, synapseManagedVirtualNetworkCostComponent(region, "Managed virtual network"))
	}

	var serverlessSqlPoolSize *decimal.Decimal
	if u != nil && u.Get("serverless_sql_pool_size_tb").Type != gjson.Null {
		serverlessSqlPoolSize = decimalPtr(decimal.NewFromInt(u.Get("serverless_sql_pool_size_tb").Int()))
	}

	if serverlessSqlPoolSize == nil || (serverlessSqlPoolSize != nil && serverlessSqlPoolSize.LessThanOrEqual(decimal.NewFromInt(10))) {
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size (first 10TB)", "0", serverlessSqlPoolSize))
	}

	if serverlessSqlPoolSize != nil && serverlessSqlPoolSize.GreaterThan(decimal.NewFromInt(10)) {
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size (first 10TB)", "0", decimalPtr(decimal.NewFromInt(10))))
		costComponents = append(costComponents, synapseServerlessSqlPoolCostComponent(region, "Serverless SQL pool size (over 10TB)", "10", decimalPtr(serverlessSqlPoolSize.Sub(decimal.NewFromInt(10)))))
	}

	var dataflowGeneralPurposeInstances, dataflowGeneralPurposeVCores *decimal.Decimal
	if u != nil && u.Get("dataflow_general_purpose_instances").Type != gjson.Null && u.Get("dataflow_general_purpose_vcores").Type != gjson.Null {
		dataflowGeneralPurposeInstances = decimalPtr(decimal.NewFromInt(u.Get("dataflow_general_purpose_instances").Int()))
		dataflowGeneralPurposeVCores = decimalPtr(decimal.NewFromInt(u.Get("dataflow_general_purpose_vcores").Int()))
	}
	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - general purpose", "DZH318Z0CGNQ", dataflowGeneralPurposeInstances, dataflowGeneralPurposeVCores))

	var dataflowComputeOptimizedInstances, dataflowComputeOptimizedVCores *decimal.Decimal
	if u != nil && u.Get("dataflow_compute_optimized_instances").Type != gjson.Null && u.Get("dataflow_compute_optimized_vcores").Type != gjson.Null {
		dataflowComputeOptimizedInstances = decimalPtr(decimal.NewFromInt(u.Get("dataflow_compute_optimized_instances").Int()))
		dataflowComputeOptimizedVCores = decimalPtr(decimal.NewFromInt(u.Get("dataflow_compute_optimized_vcores").Int()))
	}
	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - compute optimized", "DZH318Z0CGNF", dataflowComputeOptimizedInstances, dataflowComputeOptimizedVCores))

	var dataflowMemoryOptimizedInstances, dataflowMemoryOptimizedVCores *decimal.Decimal
	if u != nil && u.Get("dataflow_memory_optimized_instances").Type != gjson.Null && u.Get("dataflow_memory_optimized_vcores").Type != gjson.Null {
		dataflowMemoryOptimizedInstances = decimalPtr(decimal.NewFromInt(u.Get("dataflow_memory_optimized_instances").Int()))
		dataflowMemoryOptimizedVCores = decimalPtr(decimal.NewFromInt(u.Get("dataflow_memory_optimized_vcores").Int()))
	}
	costComponents = append(costComponents, synapseDataFlowCostComponent(region, "Data flow - memory optimized", "DZH318Z0CGNM", dataflowMemoryOptimizedInstances, dataflowMemoryOptimizedVCores))

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

func synapseManagedVirtualNetworkCostComponent(region, name string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Synapse Analytics Managed VNET")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDataFlowCostComponent(region, name, productId string, instances, vCores *decimal.Decimal) *schema.CostComponent {

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
