package google

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// CloudRunV2Job represents a resource that references a container image which is run to completion.
type CloudRunV2Job struct {
	Address              string
	Region               string
	CpuLimit             int64
	MemoryLimit          int64
	TaskCount            int64
	MonthlyJobExecutions *int64   `infracost_usage:"monthly_job_executions"`
	AvgTaskExecutionMins *float64 `infracost_usage:"average_task_execution_mins"`
}

// CoreType returns the name of this resource type
func (r *CloudRunV2Job) CoreType() string {
	return "CloudRunV2Job"
}

// UsageSchema defines a list which represents the usage schema of CloudRunV2Job.
func (r *CloudRunV2Job) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_job_executions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "average_task_execution_mins", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *CloudRunV2Job) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudRunV2Job) BuildResource() *schema.Resource {
	regionTier := GetRegionTier(r.Region)
	cpuName := "CPU allocation time"
	memoryName := "Memory allocation time"
	if regionTier == "Tier 2" {
		cpuName = "CPU allocation time (tier 2)"
		memoryName = "Memory allocation time (tier 2)"
	}

	costComponents := []*schema.CostComponent{
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
					{Key: "description", Value: strPtr(fmt.Sprintf("Jobs CPU in %s", r.Region))},
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
					{Key: "description", Value: strPtr(fmt.Sprintf("Jobs Memory in %s", r.Region))},
				},
			},
		},
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudRunV2Job) calculateCpuSeconds() *decimal.Decimal {
	if r.AvgTaskExecutionMins == nil || r.MonthlyJobExecutions == nil {
		return nil
	}

	seconds := decimal.NewFromFloat(*r.AvgTaskExecutionMins * 60)
	return decimalPtr(decimal.NewFromInt(*r.MonthlyJobExecutions).Mul(decimal.NewFromInt(r.TaskCount)).Mul(seconds).Mul(decimal.NewFromInt(r.CpuLimit)))
}

func (r *CloudRunV2Job) calculateGBSeconds() *decimal.Decimal {
	if r.AvgTaskExecutionMins == nil || r.MonthlyJobExecutions == nil {
		return nil
	}

	seconds := decimal.NewFromFloat(*r.AvgTaskExecutionMins * 60)
	gb := decimal.NewFromInt(r.MemoryLimit).Div(decimal.NewFromInt(1024 * 1024 * 1024))
	return decimalPtr(decimal.NewFromInt(*r.MonthlyJobExecutions).Mul(decimal.NewFromInt(r.TaskCount)).Mul(seconds).Mul(gb))
}
