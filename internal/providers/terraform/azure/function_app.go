package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAppFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_function_app",
		RFunc: NewAzureRMAppFunction,
		ReferenceAttributes: []string{
			"app_service_plan_id",
		},
	}
}

func NewAzureRMAppFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var memorySize *decimal.Decimal
	var instances *decimal.Decimal
	var executionTime *decimal.Decimal
	var executions *decimal.Decimal
	var skuMemory *decimal.Decimal
	var skuCPU *decimal.Decimal
	var instMulCPU *decimal.Decimal
	var instMulMemory *decimal.Decimal
	var execTimeMulMemorySize *decimal.Decimal

	var multiplicationForExecTime decimal.Decimal
	var multiplicationForCPU decimal.Decimal
	var multiplicationForMemory decimal.Decimal

	kind := "Windows"
	location := d.Get("location").String()

	if u != nil && u.Get("monthly_executions").Type != gjson.Null {
		executions = decimalPtr(decimal.NewFromFloat(u.Get("monthly_executions").Float()))
	}
	if u != nil && u.Get("execution_duration_ms").Type != gjson.Null && u.Get("memory_mb").Type != gjson.Null {
		memorySize = decimalPtr(decimal.NewFromFloat(u.Get("memory_mb").Float() / 1000))
		executionTime = decimalPtr(decimal.NewFromFloat(u.Get("execution_duration_ms").Float() / 1000))
		multiplicationForExecTime = executionTime.Mul(*memorySize)
		execTimeMulMemorySize = &multiplicationForExecTime
	}

	skuMapCPU := map[string]int64{
		"EP1": 1,
		"EP2": 2,
		"EP3": 4,
	}
	skuMapMemory := map[string]float64{
		"EP1": 3.5,
		"EP2": 7.0,
		"EP3": 14.0,
	}

	appServicePlanID := d.References("app_service_plan_id")
	skuTier := strings.ToLower(appServicePlanID[0].Get("sku.0.tier").String())
	skuSize := appServicePlanID[0].Get("sku.0.size").String()

	if len(appServicePlanID) > 0 {
		kind = strings.ToLower(appServicePlanID[0].Get("kind").String())
	}

	if val, ok := skuMapCPU[skuSize]; ok {
		skuCPU = decimalPtr(decimal.NewFromInt(val))
	}
	if val, ok := skuMapMemory[skuSize]; ok {
		skuMemory = decimalPtr(decimal.NewFromFloat(val))
	}

	if u != nil && u.Get("instances").Type != gjson.Null {
		instances = decimalPtr(decimal.NewFromFloat(u.Get("instances").Float()))
		multiplicationForCPU = instances.Mul(*skuCPU)
		multiplicationForMemory = instances.Mul(*skuMemory)
		instMulCPU = &multiplicationForCPU
		instMulMemory = &multiplicationForMemory
	}

	costComponents := make([]*schema.CostComponent, 0)

	if kind == "elastic" || skuTier == "elasticpremium" {
		costComponents = append(costComponents, AppFunctionPremiumCPUCostComponent(instMulCPU, location))
		costComponents = append(costComponents, AppFunctionPremiumMemoryCostComponent(instMulMemory, location))
	}
	if kind == "functionapp" {
		costComponents = append(costComponents, AppFunctionConsumptionExecutionTimeCostComponent(execTimeMulMemorySize, location))
		costComponents = append(costComponents, AppFunctionConsumptionExecutionsCostComponent(executions, location))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func AppFunctionPremiumCPUCostComponent(instMulCPU *decimal.Decimal, location string) *schema.CostComponent {

	return &schema.CostComponent{

		Name:           "vCPU",
		Unit:           "vCPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: instMulCPU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("vCPU Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}

}
func AppFunctionPremiumMemoryCostComponent(instMulMemory *decimal.Decimal, location string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Memory",
		Unit:           "GB-hours",
		UnitMultiplier: 1,
		HourlyQuantity: instMulMemory,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Memory Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func AppFunctionConsumptionExecutionTimeCostComponent(execTimeMulMemorySize *decimal.Decimal, location string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Execution time",
		Unit:           "GB-Seconds",
		UnitMultiplier: 1,
		HourlyQuantity: execTimeMulMemorySize,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Execution Time")},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("400000"),
		},
	}
}
func AppFunctionConsumptionExecutionsCostComponent(executions *decimal.Decimal, location string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Executions",
		Unit:            "1M requests",
		UnitMultiplier:  100000,
		MonthlyQuantity: executions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Total Executions")},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("100000"),
		},
	}
}
