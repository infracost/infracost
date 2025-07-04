package google

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// CloudRunService acts as a top-level container that manages a set of configurations and revision
// templates which implement a network service. Service exists to provide a singular abstraction which can
// be access controlled, reasoned about, and which encapsulates software lifecycle decisions such as rollout
// policy and team resource ownership.
//
// Resource information: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service/
type CloudRunService struct {
	Address                       string
	Region                        string
	CpuLimit                      int64
	IsThrottlingEnabled           bool
	MemoryLimit                   int64
	MinInstanceCount              float64
	MonthlyRequests               *int64 `infracost_usage:"monthly_requests"`
	AverageRequestDurationMs      *int64 `infracost_usage:"average_request_duration_ms"`
	ConcurrentRequestsPerInstance *int64 `infracost_usage:"concurrent_requests_per_instance"`
	InstanceHrs                   *int64 `infracost_usage:"instance_hrs"`
}

func (r *CloudRunService) CoreType() string {
	return "CloudRunService"
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
	cpuName := "CPU allocation Time"
	cpuDesc := "Services CPU (Instance-based billing) in " + r.Region
	memoryName := "Memory allocation time"
	memoryDesc := "Services Memory (Instance-based billing) in " + r.Region

	if regionTier == "Tier 2" {
		cpuName = "CPU allocation time (tier 2)"
		cpuDesc = "Services CPU Tier 2  (Request-based billing)"
		memoryName = "Memory allocation time (tier 2)"
		memoryDesc = "Services Memory Tier 2 (Request-based billing)"
	}

	var costComponents []*schema.CostComponent
	if r.IsThrottlingEnabled {
		costComponents = r.throttlingEnabledCostComponents(cpuName, cpuDesc, memoryName, memoryDesc)
	} else {
		costComponents = r.throttlingDisabledCostComponents(cpuName, memoryName)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudRunService) throttlingEnabledCostComponents(cpuName, cpuDesc, memoryName, memoryDesc string) []*schema.CostComponent {
	var requests *decimal.Decimal
	if r.MonthlyRequests != nil {
		requests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	return []*schema.CostComponent{
		{
			Name:            cpuName,
			Unit:            "vCPU-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.calculateCpuSeconds(),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(cpuDesc)},
				},
			},
		},
		{
			Name:            memoryName,
			Unit:            "GiB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.calculateGBSeconds(),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(memoryDesc)},
				},
			},
		},
		{
			Name:            "Number of requests",
			Unit:            "requests",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: requests,
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
func (r *CloudRunService) throttlingDisabledCostComponents(cpuName, memoryName string) []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            cpuName,
			Unit:            "vCPU-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.calculateCpuSeconds(),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(fmt.Sprintf("Services CPU (Instance-based billing) in %s", r.Region))},
				},
			},
		},
		{
			Name:            memoryName,
			Unit:            "GiB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.calculateGBSeconds(),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Run"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(fmt.Sprintf("Services Memory (Instance-based billing) in %s", r.Region))},
				},
			},
		},
	}
}

func (r *CloudRunService) calculateCpuSeconds() *decimal.Decimal {
	if r.IsThrottlingEnabled {
		if r.AverageRequestDurationMs == nil || r.MonthlyRequests == nil || r.ConcurrentRequestsPerInstance == nil {
			return nil
		}

		requestDurationInSeconds := decimal.NewFromInt(*r.AverageRequestDurationMs).Div(decimal.NewFromInt(1000))
		return decimalPtr(decimal.NewFromInt(*r.MonthlyRequests).Mul(requestDurationInSeconds).Div(decimal.NewFromInt(*r.ConcurrentRequestsPerInstance)).Mul(decimal.NewFromInt(r.CpuLimit)))
	}

	if r.InstanceHrs != nil && *r.InstanceHrs > 0 {
		return decimalPtr(decimal.NewFromInt(*r.InstanceHrs * 60 * 60).Mul(decimal.NewFromInt(r.CpuLimit)).Mul(decimal.NewFromFloat(r.MinInstanceCount)))
	}

	return decimalPtr(decimal.NewFromFloat(r.MinInstanceCount * (730 * 60 * 60)).Mul(decimal.NewFromInt(r.CpuLimit)))
}

func (r *CloudRunService) calculateGBSeconds() *decimal.Decimal {
	gb := decimal.NewFromInt(r.MemoryLimit).Div(decimal.NewFromInt(1024 * 1024 * 1024))
	if r.IsThrottlingEnabled {
		if r.AverageRequestDurationMs == nil || r.MonthlyRequests == nil || r.ConcurrentRequestsPerInstance == nil {
			return nil
		}

		requestDurationInSeconds := decimal.NewFromInt(*r.AverageRequestDurationMs).Div(decimal.NewFromInt(1000))
		return decimalPtr(decimal.NewFromInt(*r.MonthlyRequests).Mul(requestDurationInSeconds).Div(decimal.NewFromInt(*r.ConcurrentRequestsPerInstance)).Mul(gb))
	}

	if r.InstanceHrs != nil && *r.InstanceHrs > 0 {
		return decimalPtr(decimal.NewFromInt(*r.InstanceHrs * 60 * 60).Mul(gb).Mul(decimal.NewFromFloat(r.MinInstanceCount)))
	}

	return decimalPtr(decimal.NewFromFloat(r.MinInstanceCount * (730 * 60 * 60)).Mul(gb))
}
