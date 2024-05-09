package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// Resource information: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_service
// Pricing information: https://cloud.google.com/run/pricing
type CloudRunService struct {
	Address string
	Region  string
	CpuLimit     int64
	CpuMinScale  float64
	CpuThrottlingEnabled bool
	MemoryLimit                   int64
	MonthlyRequests     *int64 `infracost_usage:"monthly_requests"`
	AverageRequestDurationMs *int64 `infracost_usage:"average_request_duration_ms"`
	ConcurrentRequestsPerInstance  *int64 `infracost_usage:"concurrent_requests_per_instance"`
	InstanceHrs         *int64 `infracost_usage:"instance_hrs"`
}

func (r *CloudRunService) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "average_request_duration_ms", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "concurrent_requests_per_instance", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "instance_hrs", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *CloudRunService) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudRunService) BuildResource() *schema.Resource {
	regionTier := GetRegionTier(r.Region)
	var cpuName string
	var memoryName string
	if regionTier == "Tier 2" {
		cpuName  = "CPU Allocation Time (tier 2)"
		memoryName = "Memory Allocation Time (tier 2)"
	} else {
		cpuName  = "CPU Allocation Time"
		memoryName = "Memory Allocation Time"
	}
	var costComponents []*schema.CostComponent
	if r.CpuThrottlingEnabled {
		costComponents = r.throttlingEnabledCostComponent(cpuName, memoryName)
	} else {
		costComponents = r.throttlingDisabledCostComponent()
	}
	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudRunService) throttlingEnabledCostComponent(cpuName string, memoryName string) []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            cpuName,
			Unit:            "vCPU-seconds",
			UnitMultiplier:  decimal.NewFromInt(1), 
			MonthlyQuantity: decimalPtr(r.calculateCpuSeconds(true)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(cpuName)},
				},
			},
		},
		{
			Name:            memoryName,
			Unit:            "GiB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1), 
			MonthlyQuantity: decimalPtr(r.calculateGBSeconds(true)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(memoryName)},
				},
			},
		},
		{
			Name:            "Number of requests",
			Unit:            "request",
			UnitMultiplier:  decimal.NewFromInt(1), 
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.MonthlyRequests)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr("Requests")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("2000000"),
			},
		},
		
	}
}
func (r *CloudRunService) throttlingDisabledCostComponent() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "CPU Allocation Time (Always-on)",
			Unit:            "vCPU-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(r.calculateCpuSeconds(false)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(fmt.Sprintf("CPU Allocation Time (Always-on CPU) in %s", r.Region))},
				},
			},
		},
		{
			Name:            "Memory Allocation Time (Always-on)",
			Unit:            "GiB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1), 
			MonthlyQuantity: decimalPtr(r.calculateGBSeconds(false)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(fmt.Sprintf("Memory Allocation Time (Always-on CPU) in %s", r.Region))},
				},
			},
		},
	}
}

func (r *CloudRunService) calculateCpuSeconds(throttlingEnabled bool) decimal.Decimal {
	var cpuSeconds decimal.Decimal
	if throttlingEnabled {
		requestDurationInSeconds := decimal.NewFromInt(*r.AverageRequestDurationMs).Div(decimal.NewFromInt(1000))
		cpuSeconds = decimal.NewFromInt(*r.MonthlyRequests).Mul(requestDurationInSeconds).Div(decimal.NewFromInt(*r.ConcurrentRequestsPerInstance)).Mul(decimal.NewFromInt(r.CpuLimit))
	} else {
		if r.InstanceHrs != nil && *r.InstanceHrs > 0 {
			cpuSeconds = decimal.NewFromInt(*r.InstanceHrs * 60 * 60).Mul(decimal.NewFromInt(r.CpuLimit)).Mul(decimal.NewFromFloat(r.CpuMinScale))
		} else {
			cpuSeconds = decimal.NewFromFloat(r.CpuMinScale * (730 * 60 * 60)).Mul(decimal.NewFromInt(r.CpuLimit))
		}
	}
	return cpuSeconds
}

func (r *CloudRunService) calculateGBSeconds(throttlingEnabled bool) decimal.Decimal {
	var seconds decimal.Decimal
	gb := decimal.NewFromInt(r.MemoryLimit).Div(decimal.NewFromInt(1024 * 1024 * 1024))
	if throttlingEnabled {
		requestDurationInSeconds := decimal.NewFromInt(*r.AverageRequestDurationMs).Div(decimal.NewFromInt(1000))
		seconds = decimal.NewFromInt(*r.MonthlyRequests).Mul(requestDurationInSeconds).Div(decimal.NewFromInt(*r.ConcurrentRequestsPerInstance)).Mul(gb)
	} else {
		if r.InstanceHrs != nil && *r.InstanceHrs > 0 {
			seconds = decimal.NewFromInt(*r.InstanceHrs * 60 * 60).Mul(gb).Mul(decimal.NewFromFloat(r.CpuMinScale))
		} else {
			seconds = decimal.NewFromFloat(r.CpuMinScale * (730 * 60 * 60)).Mul(gb)
		}
	}
	return seconds
}