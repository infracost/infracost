package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
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
	region := d.Region

	costComponents := make([]*schema.CostComponent, 0)

	managedVirtualNetwork := false
	if d.Get("managed_virtual_network_enabled").Type != gjson.Null {
		managedVirtualNetwork = d.Get("managed_virtual_network_enabled").Bool()
	}

	var serverlessSQLPoolSize *decimal.Decimal
	if u != nil && u.Get("serverless_sql_pool_size_tb").Type != gjson.Null {
		serverlessSQLPoolSize = decimalPtr(decimal.NewFromInt(u.Get("serverless_sql_pool_size_tb").Int()))
	}

	costComponents = append(costComponents, synapseServerlessSQLPoolCostComponent(region, "Serverless SQL pool size", "10", serverlessSQLPoolSize))

	dataflowTiers := [2]string{"Basic", "Standard"}
	for _, tier := range dataflowTiers {
		var dataflowInstances, dataflowVCores, dataflowHours *decimal.Decimal

		var instancesUsageKey = fmt.Sprintf("dataflow_%s_instances", strings.ToLower(tier))
		var vcoresUsageKey = fmt.Sprintf("dataflow_%s_vcores", strings.ToLower(tier))
		var hoursUsageKey = fmt.Sprintf("monthly_dataflow_%s_hours", strings.ToLower(tier))

		if u != nil && u.Get(instancesUsageKey).Type != gjson.Null && u.Get(vcoresUsageKey).Type != gjson.Null && u.Get(hoursUsageKey).Type != gjson.Null {
			dataflowInstances = decimalPtr(decimal.NewFromInt(u.Get(instancesUsageKey).Int()))
			dataflowVCores = decimalPtr(decimal.NewFromInt(u.Get(vcoresUsageKey).Int()))
			dataflowHours = decimalPtr(decimal.NewFromInt(u.Get(hoursUsageKey).Int()))
		}
		costComponents = append(costComponents, synapseDataFlowCostComponent(region, fmt.Sprintf("Data flow (%s)", strings.ToLower(tier)), tier, dataflowInstances, dataflowVCores, dataflowHours))
	}

	datapipelineTiers := [2]string{"Azure Hosted IR", "Self Hosted IR"}
	datapipelineUsageKeys := [2]string{"azure_hosted", "self_hosted"}
	if managedVirtualNetwork {
		datapipelineTiers = [2]string{"Azure Hosted Managed VNET IR", "Self Hosted IR"}
	}

	for i, tier := range datapipelineTiers {
		var activityRuns, dataIntegrationUnits, dataIntegrationHours, dataMovementHours, integrationRuntimeHours, externalIntegrationRuntimeHours *decimal.Decimal
		var usageName = strings.Replace(datapipelineUsageKeys[i], "_", " ", 1)

		var activityRunsUsageKey = fmt.Sprintf("monthly_datapipeline_%s_activity_runs", datapipelineUsageKeys[i])
		if u != nil && u.Get(activityRunsUsageKey).Type != gjson.Null {
			activityRuns = decimalPtr(decimal.NewFromInt(u.Get(activityRunsUsageKey).Int()))
		}
		costComponents = append(costComponents, synapseDataPipelineActivityRunCostComponent(region, fmt.Sprintf("Data pipeline %s activity runs", usageName), tier, "Orchestration Activity Run", activityRuns))

		if datapipelineUsageKeys[i] == "azure_hosted" {
			var dataIntegrationUnitUsageKey = fmt.Sprintf("monthly_datapipeline_%s_data_integration_units", datapipelineUsageKeys[i])
			var dataIntegrationHoursUsageKey = fmt.Sprintf("monthly_datapipeline_%s_data_integration_hours", datapipelineUsageKeys[i])
			if u != nil && u.Get(dataIntegrationUnitUsageKey).Type != gjson.Null && u.Get(dataIntegrationHoursUsageKey).Type != gjson.Null {
				dataIntegrationUnits = decimalPtr(decimal.NewFromInt(u.Get(dataIntegrationUnitUsageKey).Int()))
				dataIntegrationHours = decimalPtr(decimal.NewFromInt(u.Get(dataIntegrationHoursUsageKey).Int()))
			}
			costComponents = append(costComponents, synapseDataPipelineDataMovementCostComponent(region, fmt.Sprintf("Data pipeline %s data integration units", usageName), tier, "Data Movement", "DIU-hours", dataIntegrationUnits, dataIntegrationHours))
		} else {
			var dataMovementHoursUsageKey = fmt.Sprintf("monthly_datapipeline_%s_data_movement_hours", datapipelineUsageKeys[i])
			if u != nil && u.Get(dataMovementHoursUsageKey).Type != gjson.Null {
				dataMovementHours = decimalPtr(decimal.NewFromInt(u.Get(dataMovementHoursUsageKey).Int()))
			}
			costComponents = append(costComponents, synapseDataPipelineDataMovementCostComponent(region, fmt.Sprintf("Data pipeline %s data movement", usageName), tier, "Data Movement", "hours", decimalPtr(decimal.NewFromInt(1)), dataMovementHours))
		}

		var integrationRuntimeUsageKey = fmt.Sprintf("monthly_datapipeline_%s_integration_runtime_hours", datapipelineUsageKeys[i])
		if u != nil && u.Get(integrationRuntimeUsageKey).Type != gjson.Null {
			integrationRuntimeHours = decimalPtr(decimal.NewFromInt(u.Get(integrationRuntimeUsageKey).Int()))
		}
		costComponents = append(costComponents, synapseDataPipelineActivityIntegrationRuntimeCostComponent(region, fmt.Sprintf("Data pipeline %s integration runtime", usageName), tier, "Pipeline Activity", integrationRuntimeHours))

		var externalIntegrationRuntimeUsageKey = fmt.Sprintf("monthly_datapipeline_%s_external_integration_runtime_hours", datapipelineUsageKeys[i])
		if u != nil && u.Get(externalIntegrationRuntimeUsageKey).Type != gjson.Null {
			externalIntegrationRuntimeHours = decimalPtr(decimal.NewFromInt(u.Get(externalIntegrationRuntimeUsageKey).Int()))
		}
		costComponents = append(costComponents, synapseDataPipelineActivityIntegrationRuntimeCostComponent(region, fmt.Sprintf("Data pipeline %s external integration runtime", usageName), tier, "External Pipeline Activity", externalIntegrationRuntimeHours))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func synapseServerlessSQLPoolCostComponent(region, name, start string, quantity *decimal.Decimal) *schema.CostComponent {
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

func synapseDataPipelineActivityRunCostComponent(region, name, sku, meter string, runs *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "1k activity runs",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: runs,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s %s", sku, meter))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDataPipelineDataMovementCostComponent(region, name, sku, meter, unit string, diu, hours *decimal.Decimal) *schema.CostComponent {

	var hourlyQuantity *decimal.Decimal
	if diu != nil && hours != nil {
		hourlyQuantity = decimalPtr(diu.Mul(*hours))
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: hourlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s %s", sku, meter))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDataPipelineActivityIntegrationRuntimeCostComponent(region, name, sku, meter string, hours *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: hours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s %s", sku, meter))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDataFlowCostComponent(region, name, tier string, instances, vCores, hours *decimal.Decimal) *schema.CostComponent {

	var hourlyQuantity *decimal.Decimal
	if instances != nil && vCores != nil && hours != nil {
		hourlyQuantity = decimalPtr(vCores.Mul(*instances).Mul(*hours))
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "vCore-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: hourlyQuantity,
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
