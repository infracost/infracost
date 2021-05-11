package azure

import (
	"fmt"
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
	meterNameCPU := "vCPU Duration"
	meterNameMemory := "Memory Duration"
	meterNameExecTime := "Execution Time"
	meterNameExecutions := "Total Executions"
	memorySize := 1.0
	executionTime := 1.1
	executions := 1.1
	kind := "kind"
	skuTier := "tier"
	skuSize := "EP"
	skuMemory := 3.0
	skuCPU := 1
	instances := 1.0
	location := "someus"

	if d.Get("location").Type != gjson.Null {
		location = d.Get("location").String()
	}
	if u != nil && u.Get("monthly_executions").Type != gjson.Null {
		executions = u.Get("monthly_executions").Float()
	}
	if u != nil && u.Get("execution_duration_ms").Type != gjson.Null {
		executionTime = u.Get("execution_duration_ms").Float() / 1000
		fmt.Println(executionTime)
	}
	if u != nil && u.Get("memory_mb").Type != gjson.Null {
		memorySize = u.Get("memory_mb").Float() / 1000
	}

	skuMapCPU := map[string]int{
		"EP1": 1,
		"EP2": 2,
		"EP3": 4,
	}
	skuMapMemory := map[string]float64{
		"EP1": 3.5,
		"EP2": 7.0,
		"EP3": 14.0,
	}

	app_service_plan_id := d.References("app_service_plan_id")

	if len(app_service_plan_id) > 0 {
		kind = strings.ToLower(app_service_plan_id[0].Get("kind").String())
		skuTier = strings.ToLower(app_service_plan_id[0].Get("sku.0.tier").String())
	}
	if app_service_plan_id[0].Get("sku.0.size").Type != gjson.Null {
		skuSize = app_service_plan_id[0].Get("sku.0.size").String()
	}

	if val, ok := skuMapCPU[skuSize]; ok {
		skuCPU = val
	}
	if val, ok := skuMapMemory[skuSize]; ok {
		skuMemory = val
	}

	if u != nil && u.Get("instances").Type != gjson.Null {
		instances = u.Get("instances").Float()
	}

	costComponents := make([]*schema.CostComponent, 0)

	if kind == "elastic" || skuTier == "elasticpremium" {
		costComponents = append(costComponents, AppFunctionPremiumCPUCostComponent(instances, skuCPU, location, meterNameCPU))
		costComponents = append(costComponents, AppFunctionPremiumMemoryCostComponent(instances, skuMemory, location, meterNameMemory))
	}
	if kind == "functionapp" {
		costComponents = append(costComponents, AppFunctionConsumptionExecutionTimeCostComponent(executions, executionTime, memorySize, location, meterNameExecTime))
		costComponents = append(costComponents, AppFunctionConsumptionExecutionsCostComponent(executions, location, meterNameExecutions))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func AppFunctionPremiumCPUCostComponent(instances float64, skuCPU int, location, meterName string) *schema.CostComponent {

	return &schema.CostComponent{

		Name:           "vCPU",
		Unit:           "vCPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromFloat(instances * float64(skuCPU))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}

}
func AppFunctionPremiumMemoryCostComponent(instances, skuMemory float64, location, meterName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Memory",
		Unit:           "GB-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromFloat(instances * skuMemory)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func AppFunctionConsumptionExecutionTimeCostComponent(executions, executionTime float64, memorySize float64, location, meterName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Execution time",
		Unit:           "GB Seconds",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromFloat(executionTime * memorySize)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(meterName)},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func AppFunctionConsumptionExecutionsCostComponent(executions float64, location, meterName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Executions",
		Unit:            "requests",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(executions / 10)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(meterName)},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
