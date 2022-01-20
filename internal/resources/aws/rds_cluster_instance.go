package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"strings"

	"github.com/shopspring/decimal"
)

type RDSClusterInstance struct {
	Address             string
	Region              string
	InstanceClass       string
	Engine              string
	MonthlyCPUCreditHrs *int64 `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount           *int64 `infracost_usage:"vcpu_count"`
}

var RDSClusterInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "vcpu_count", ValueType: schema.Int64, DefaultValue: 0},
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

	if strings.HasPrefix(r.InstanceClass, "db.t3") {
		instanceCPUCreditHours := decimal.Zero
		if r.MonthlyCPUCreditHrs != nil {
			instanceCPUCreditHours = decimal.NewFromInt(*r.MonthlyCPUCreditHrs)
		}

		instanceVCPUCount := decimal.Zero
		if r.VCPUCount != nil {
			instanceVCPUCount = decimal.NewFromInt(*r.VCPUCount)
		}

		if instanceCPUCreditHours.GreaterThan(decimal.NewFromInt(0)) {
			cpuCreditQuantity := instanceVCPUCount.Mul(instanceCPUCreditHours)
			costComponents = append(costComponents, r.cpuCreditsCostComponent(databaseEngine, cpuCreditQuantity))
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

func (r *RDSClusterInstance) cpuCreditsCostComponent(databaseEngine string, vCPUCount decimal.Decimal) *schema.CostComponent {
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
			},
		},
	}
}
