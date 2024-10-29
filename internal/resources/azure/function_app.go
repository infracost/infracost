package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

var (
	functionAppSkuMapCPU = map[string]int64{
		"ep1": 1,
		"ep2": 2,
		"ep3": 4,
	}

	functionAppSkuMapMem = map[string]float64{
		"ep1": 3.5,
		"ep2": 7.0,
		"ep3": 14.0,
	}
)

// FunctionApp struct a serverless function running in an app service environment. The billing for this
// function lies within Azure App Service, however we capture the costs in this component to make it more understandable.
//
// Resource information: https://learn.microsoft.com/en-us/azure/azure-functions/functions-overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/app-service/windows/
type FunctionApp struct {
	Address string
	Region  string

	SKUName string
	Tier    string
	OSType  string

	MonthlyExecutions   *int64 `infracost_usage:"monthly_executions"`
	ExecutionDurationMs *int64 `infracost_usage:"execution_duration_ms"`
	MemoryMb            *int64 `infracost_usage:"memory_mb"`
	Instances           *int64 `infracost_usage:"instances"`
}

func (r *FunctionApp) CoreType() string {
	return "FunctionApp"
}

func (r *FunctionApp) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_executions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "execution_duration_ms", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "memory_mb", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "instances", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the FunctionApp struct
// It uses the `infracost_usage` struct tags to populate data into the FunctionApp
func (r *FunctionApp) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid FunctionApp struct.
//
// FunctionApp costs are CPU and Memory usage. These values rely on the user defining their expected
// usage in the usage file.
//
// Function apps are billed in two modes - Premium or Consumption.
func (r *FunctionApp) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.Tier == "premium" {
		cpu := r.appFunctionPremiumCPUCostComponent()
		if cpu != nil {
			costComponents = append(costComponents, cpu)
		}

		mem := r.appFunctionPremiumMemoryCostComponent()
		if mem != nil {
			costComponents = append(costComponents, mem)
		}

		return &schema.Resource{
			Name:           r.Address,
			CostComponents: costComponents,
			UsageSchema:    r.UsageSchema(),
		}
	}

	costComponents = append(
		costComponents,
		r.appFunctionConsumptionExecutionTimeCostComponent(),
		r.appFunctionConsumptionExecutionsCostComponent(),
	)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *FunctionApp) appFunctionPremiumCPUCostComponent() *schema.CostComponent {
	var skuCPU *int64

	if val, ok := functionAppSkuMapCPU[r.SKUName]; ok {
		skuCPU = &val
	}

	if skuCPU == nil {
		return nil
	}

	instances := decimal.NewFromInt(1)
	if r.Instances != nil {
		instances = decimal.NewFromInt(*r.Instances)
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("vCPU (%s)", strings.ToUpper(r.SKUName)),
		Unit:           "vCPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(instances.Mul(decimal.NewFromInt(*skuCPU))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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

func (r *FunctionApp) appFunctionPremiumMemoryCostComponent() *schema.CostComponent {
	var skuMemory *float64

	if val, ok := functionAppSkuMapMem[r.SKUName]; ok {
		skuMemory = &val
	}

	if skuMemory == nil {
		return nil
	}

	instances := decimal.NewFromInt(1)
	if r.Instances != nil {
		instances = decimal.NewFromInt(*r.Instances)
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Memory (%s)", strings.ToUpper(r.SKUName)),
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(instances.Mul(decimal.NewFromFloat(*skuMemory))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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

func (r *FunctionApp) appFunctionConsumptionExecutionTimeCostComponent() *schema.CostComponent {
	gbSeconds := r.calculateFunctionAppGBSeconds()
	return &schema.CostComponent{
		Name:            "Execution time",
		Unit:            "GB-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: gbSeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
		UsageBased: true,
	}
}

func (r *FunctionApp) appFunctionConsumptionExecutionsCostComponent() *schema.CostComponent {
	// Azure's pricing API returns prices per 10 executions so if the user has provided
	// the number of executions, we should divide it by 10
	var executions *decimal.Decimal
	if r.MonthlyExecutions != nil {
		executions = decimalPtr(decimal.NewFromInt(*r.MonthlyExecutions).Div(decimal.NewFromInt(10)))
	}

	return &schema.CostComponent{
		Name:            "Executions",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(100000),
		MonthlyQuantity: executions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
		UsageBased: true,
	}
}

func (r *FunctionApp) calculateFunctionAppGBSeconds() *decimal.Decimal {
	if r.MemoryMb == nil || r.ExecutionDurationMs == nil || r.MonthlyExecutions == nil {
		return nil
	}

	memorySize := decimal.NewFromInt(*r.MemoryMb)
	averageRequestDuration := decimal.NewFromInt(*r.ExecutionDurationMs)
	monthlyRequests := decimal.NewFromInt(*r.MonthlyExecutions)

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

	return &gbSeconds
}
