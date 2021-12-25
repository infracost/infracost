package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
	"strings"
)

type RDSClusterInstance struct {
	Address             *string
	Region              *string
	InstanceClass       *string
	Engine              *string
	VcpuCount           *int64 `infracost_usage:"vcpu_count"`
	MonthlyCpuCreditHrs *int64 `infracost_usage:"monthly_cpu_credit_hrs"`
}

var RDSClusterInstanceUsageSchema = []*schema.UsageItem{{Key: "vcpu_count", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0}}

func (r *RDSClusterInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RDSClusterInstance) BuildResource() *schema.Resource {
	region := *r.Region

	instanceType := *r.InstanceClass

	var databaseEngine *string
	switch *r.Engine {
	case "aurora", "aurora-mysql", "":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (%s, %s)", "on-demand", instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
					{Key: "databaseEngine", Value: databaseEngine},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if strings.HasPrefix(instanceType, "db.t3") {
		instanceCPUCreditHours := decimal.Zero
		if r != nil && r.MonthlyCpuCreditHrs != nil {
			instanceCPUCreditHours = decimal.NewFromInt(*r.MonthlyCpuCreditHrs)
		}

		instanceVCPUCount := decimal.Zero
		if r != nil && r.VcpuCount != nil {
			instanceVCPUCount = decimal.NewFromInt(*r.VcpuCount)
		}

		if instanceCPUCreditHours.GreaterThan(decimal.NewFromInt(0)) {
			cpuCreditQuantity := instanceVCPUCount.Mul(instanceCPUCreditHours)
			costComponents = append(costComponents, rdsCPUCreditsCostComponent(region, databaseEngine, cpuCreditQuantity))
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: RDSClusterInstanceUsageSchema,
	}
}

func rdsCPUCreditsCostComponent(region string, databaseEngine *string, vCPUCount decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "CPU credits",
		Unit:            "vCPU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &vCPUCount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("CPU Credits"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", Value: databaseEngine},
			},
		},
	}
}
