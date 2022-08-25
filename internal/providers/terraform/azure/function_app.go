package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
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
	region := lookupRegion(d, []string{})

	var memorySize, executionTime, executions, gbSeconds *decimal.Decimal
	var skuCPU *int64
	var skuMemory *float64
	var skuTier, skuSize string
	kind := "Windows"

	if u != nil && u.Get("monthly_executions").Type != gjson.Null {
		executions = decimalPtr(decimal.NewFromInt(u.Get("monthly_executions").Int()))
	}
	if u != nil && u.Get("execution_duration_ms").Type != gjson.Null &&
		u.Get("memory_mb").Type != gjson.Null &&
		executions != nil {

		memorySize = decimalPtr(decimal.NewFromInt(u.Get("memory_mb").Int()))
		executionTime = decimalPtr(decimal.NewFromInt(u.Get("execution_duration_ms").Int()))
		gbSeconds = decimalPtr(calculateFunctionAppGBSeconds(*memorySize, *executionTime, *executions))
	}

	skuMapCPU := map[string]int64{
		"ep1": 1,
		"ep2": 2,
		"ep3": 4,
	}
	skuMapMemory := map[string]float64{
		"ep1": 3.5,
		"ep2": 7.0,
		"ep3": 14.0,
	}

	appServicePlanID := d.References("app_service_plan_id")

	if len(appServicePlanID) > 0 {
		skuTier = strings.ToLower(appServicePlanID[0].Get("sku.0.tier").String())
		skuSize = strings.ToLower(appServicePlanID[0].Get("sku.0.size").String())
		kind = strings.ToLower(appServicePlanID[0].Get("kind").String())
	}

	if val, ok := skuMapCPU[skuSize]; ok {
		skuCPU = &val
	}
	if val, ok := skuMapMemory[skuSize]; ok {
		skuMemory = &val
	}

	instances := decimal.NewFromInt(1)
	if u != nil && u.Get("instances").Type != gjson.Null {
		instances = decimal.NewFromInt(u.Get("instances").Int())
	}

	costComponents := make([]*schema.CostComponent, 0)

	if (strings.ToLower(kind) == "elastic" || strings.ToLower(skuTier) == "elasticpremium") && skuCPU != nil && skuMemory != nil {
		costComponents = append(costComponents, AppFunctionPremiumCPUCostComponent(skuSize, instances, skuCPU, region))
		costComponents = append(costComponents, AppFunctionPremiumMemoryCostComponent(skuSize, instances, skuMemory, region))
	} else {
		if strings.ToLower(kind) == "functionapp" || gbSeconds != nil {
			costComponents = append(costComponents, AppFunctionConsumptionExecutionTimeCostComponent(gbSeconds, region))
		}
		if strings.ToLower(kind) == "functionapp" || executions != nil {
			costComponents = append(costComponents, AppFunctionConsumptionExecutionsCostComponent(executions, region))
		}
	}

	if len(costComponents) > 1 {
		return &schema.Resource{
			Name:           d.Address,
			CostComponents: costComponents,
		}
	}
	log.Warnf("Skipping resource %s. Could not find a way to get its cost components from the resource or usage file.", d.Address)
	return nil
}

func AppFunctionPremiumCPUCostComponent(skuSize string, instances decimal.Decimal, skuCPU *int64, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("vCPU (%s)", strings.ToUpper(skuSize)),
		Unit:           "vCPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(instances.Mul(decimal.NewFromInt(*skuCPU))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr("vCPU Duration$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func AppFunctionPremiumMemoryCostComponent(skuSize string, instances decimal.Decimal, skuMemory *float64, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Memory (%s)", strings.ToUpper(skuSize)),
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(instances.Mul(decimal.NewFromFloat(*skuMemory))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr("Memory Duration$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func AppFunctionConsumptionExecutionTimeCostComponent(gbSeconds *decimal.Decimal, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Execution time",
		Unit:            "GB-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: gbSeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr("Execution Time$")},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("400000"),
		},
	}
}

func AppFunctionConsumptionExecutionsCostComponent(executions *decimal.Decimal, region string) *schema.CostComponent {
	// Azure's pricing API returns prices per 10 executions so if the user has provided the number of executions, we should divide it by 10
	if executions != nil {
		executions = decimalPtr(executions.Div(decimal.NewFromInt(10)))
	}

	return &schema.CostComponent{
		Name:            "Executions",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(100000),
		MonthlyQuantity: executions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Functions"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr("Total Executions$")},
				{Key: "skuName", Value: strPtr("Standard")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("100000"),
		},
	}
}

func calculateFunctionAppGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	// Use a min of 128MB, and round-up to nearest 128MB
	if memorySize.LessThan(decimal.NewFromInt(128)) {
		memorySize = decimal.NewFromInt(128)
	}
	roundedMemory := memorySize.Div(decimal.NewFromInt(128)).Ceil().Mul(decimal.NewFromInt(128))
	// Apply the minimum request duration
	if averageRequestDuration.LessThan(decimal.NewFromInt(100)) {
		averageRequestDuration = decimal.NewFromInt(100)
	}
	durationSeconds := monthlyRequests.Mul(averageRequestDuration).Mul(decimal.NewFromFloat(0.001))
	gbSeconds := durationSeconds.Mul(roundedMemory).Div(decimal.NewFromInt(1024))
	return gbSeconds
}
