package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"fmt"
)

// Resource information: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_service
// Pricing information: https://cloud.google.com/run/pricing
type CloudRunService struct {
	Address string
	Region  string
	CpuLimit     int64
	CpuMinScale  int64
	CpuThrottlingEnabled bool
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

// PopulateUsage parses the u schema.UsageData into the CloudRunService.
// It uses the `infracost_usage` struct tags to populate data into the CloudRunService.
func (r *CloudRunService) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CloudRunService struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CloudRunService) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent
	if r.CpuThrottlingEnabled {
		costComponents = []*schema.CostComponent{
			r.throttlingEnabledCostComponent(),
		}
	} else {
		costComponents = []*schema.CostComponent{
			r.throttlingDisabledCostComponent(),
		}
	}
	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudRunService) throttlingEnabledCostComponent() *schema.CostComponent {
	// Instance hours are calculated as (monthly requests * average request duration) / concurrent requests per instance
	requestDurationInSeconds := decimal.NewFromInt(*r.AverageRequestDurationMs).Div(decimal.NewFromInt(1000))
	return &schema.CostComponent{
		Name:            "CPU Allocation Time",
		Unit:            "vCPU-seconds",
		UnitMultiplier:  decimal.NewFromInt(1), // quantity/unitMultiplier * price
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.MonthlyRequests).Mul(requestDurationInSeconds).Div(decimal.NewFromInt(*r.ConcurrentRequestsPerInstance))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Run"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("CPU Allocation Time")},
			},
		},
	}
}
func (r *CloudRunService) throttlingDisabledCostComponent() *schema.CostComponent {
	// Instance hours if specified, instance_hrs * CPU limit
	// Instance hours not specified, instance_hrs = (minScale * 730) * CPU limit
	
	// var cpuUsage *decimal.Decimal
	// if r.InstanceHrs != nil && *r.InstanceHrs > 0 {
	// 	cpuUsage = decimalPtr(decimal.NewFromInt(*r.InstanceHrs).Mul(decimal.NewFromInt(r.CpuLimit)))
	// }
	// else {
	// 	cpuUsage = decimalPtr(decimal.NewFromInt(r.CpuMinScale * 730))
	// }

	return &schema.CostComponent{
		Name:            "CPU Allocation Time (Always-on)",
		Unit:            "vCPU-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: intPtrToDecimalPtr(r.MonthlyRequests),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Run"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("CPU Allocation Time (Always-on CPU) in %s", r.Region))},
			},
		},
	}
}

