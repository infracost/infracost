package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"strings"
)

type RDSClusterInstance struct {
	Address                                      string
	Region                                       string
	InstanceClass                                string
	Engine                                       string
	PerformanceInsightsEnabled                   bool
	PerformanceInsightsLongTermRetention         bool
	MonthlyCPUCreditHrs                          *int64 `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                                    *int64 `infracost_usage:"vcpu_count"`
	MonthlyAdditionalPerformanceInsightsRequests *int64 `infracost_usage:"monthly_additional_performance_insights_requests"`
}

var RDSClusterInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "vcpu_count", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "monthly_additional_performance_insights_requests", ValueType: schema.Int64, DefaultValue: 0},
}

func (r *RDSClusterInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RDSClusterInstance) BuildResource() *schema.Resource {
	databaseEngine := r.databaseEngineValue()

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (%s, %s)", "on-demand", r.InstanceClass),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(r.InstanceClass)},
					{Key: "databaseEngine", Value: strPtr(databaseEngine)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if instanceFamily := getBurstableInstanceFamily([]string{"db.t3", "db.t4g"}, r.InstanceClass); instanceFamily != "" {
		instanceCPUCreditHours := decimal.Zero
		if r.MonthlyCPUCreditHrs != nil {
			instanceCPUCreditHours = decimal.NewFromInt(*r.MonthlyCPUCreditHrs)
		}

		instanceVCPUCount := decimal.Zero
		if r.VCPUCount != nil {
			// VCPU count has been set explicitly
			instanceVCPUCount = decimal.NewFromInt(*r.VCPUCount)
		} else if count, ok := InstanceTypeToVCPU[strings.TrimPrefix(r.InstanceClass, "db.")]; ok {
			// We were able to lookup thing VCPU count
			instanceVCPUCount = decimal.NewFromInt(count)
		}

		if instanceCPUCreditHours.GreaterThan(decimal.NewFromInt(0)) {
			cpuCreditQuantity := instanceVCPUCount.Mul(instanceCPUCreditHours)
			costComponents = append(costComponents, r.cpuCreditsCostComponent(databaseEngine, instanceFamily, cpuCreditQuantity))
		}
	}
	if r.PerformanceInsightsEnabled {
		if r.PerformanceInsightsLongTermRetention {
			costComponents = append(costComponents, performanceInsightsLongTermRetentionCostComponent(r.Region, r.InstanceClass))
		}

		if r.MonthlyAdditionalPerformanceInsightsRequests == nil || *r.MonthlyAdditionalPerformanceInsightsRequests > 0 {
			costComponents = append(costComponents,
				performanceInsightsAPIRequestCostComponent(r.Region, r.MonthlyAdditionalPerformanceInsightsRequests))
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    RDSClusterInstanceUsageSchema,
	}
}

func (r *RDSClusterInstance) databaseEngineValue() string {
	switch r.Engine {
	case "aurora", "aurora-mysql", "":
		return "Aurora MySQL"
	case "aurora-postgresql":
		return "Aurora PostgreSQL"
	}

	return ""
}

func (r *RDSClusterInstance) cpuCreditsCostComponent(databaseEngine, instanceFamily string, vCPUCount decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "CPU credits",
		Unit:            "vCPU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &vCPUCount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("CPU Credits"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", Value: strPtr(databaseEngine)},
				{Key: "usagetype", ValueRegex: regexPtr("CPUCredits:" + instanceFamily + "$")},
			},
		},
	}
}
