package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type CloudScheduler struct {
	Address         string
	Region          string
	MonthlyJobCount *int64 `infracost_usage:"monthly_job_count"`
}

func (r *CloudScheduler) CoreType() string {
	return "CloudScheduler"
}

func (r *CloudScheduler) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_job_count", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *CloudScheduler) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudScheduler) BuildResource() *schema.Resource {
	var jobCount int64
	if r.MonthlyJobCount != nil {
		jobCount = *r.MonthlyJobCount
	}

	tierLimits := []int{3} // First 3 jobs are free
	tierQuantities := usage.CalculateTierBuckets( decimal.NewFromInt(jobCount), tierLimits)

	costComponents := []*schema.CostComponent{}

	if tierQuantities[1].GreaterThan(decimal.Zero) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Cloud Scheduler jobs (above free tier)",
			Unit:            "jobs",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &tierQuantities[1],
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud Scheduler"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr("CloudScheduler")},
					{Key: "description", Value: strPtr("Cloud Scheduler jobs")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("3"),
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
