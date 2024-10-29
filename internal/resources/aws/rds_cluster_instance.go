package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type RDSClusterInstance struct {
	Address                                      string
	Region                                       string
	InstanceClass                                string
	Engine                                       string
	Version                                      string
	IOOptimized                                  bool
	PerformanceInsightsEnabled                   bool
	PerformanceInsightsLongTermRetention         bool
	MonthlyCPUCreditHrs                          *int64   `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                                    *int64   `infracost_usage:"vcpu_count"`
	MonthlyAdditionalPerformanceInsightsRequests *int64   `infracost_usage:"monthly_additional_performance_insights_requests"`
	ReservedInstanceTerm                         *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption                *string  `infracost_usage:"reserved_instance_payment_option"`
	CapacityUnitsPerHr                           *float64 `infracost_usage:"capacity_units_per_hr"`
}

func (r *RDSClusterInstance) CoreType() string {
	return "RDSClusterInstance"
}

func (r *RDSClusterInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "vcpu_count", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_additional_performance_insights_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
		{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
		{Key: "capacity_units_per_hr", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *RDSClusterInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RDSClusterInstance) BuildResource() *schema.Resource {
	databaseEngine := r.databaseEngineValue()

	costComponents := []*schema.CostComponent{}
	isServerless := strings.EqualFold(r.InstanceClass, "db.serverless")
	if isServerless {
		costComponents = append(costComponents, r.auroraServerlessV2CostComponent(databaseEngine))
	} else {
		costComponents = append(costComponents, r.dbInstanceCostComponent(databaseEngine))
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
			costComponents = append(costComponents, performanceInsightsLongTermRetentionCostComponent(r.Region, r.InstanceClass, databaseEngine, isServerless, r.CapacityUnitsPerHr))
		}

		if r.MonthlyAdditionalPerformanceInsightsRequests == nil || *r.MonthlyAdditionalPerformanceInsightsRequests > 0 {
			costComponents = append(costComponents,
				performanceInsightsAPIRequestCostComponent(r.Region, r.MonthlyAdditionalPerformanceInsightsRequests))
		}
	}

	extendedSupport := extendedSupportCostComponent(r.Version, r.Region, r.Engine, r.InstanceClass)
	if extendedSupport != nil {
		costComponents = append(costComponents, extendedSupport)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *RDSClusterInstance) databaseEngineValue() string {
	if r.Engine == "aurora-postgresql" {
		return "Aurora PostgreSQL"
	}

	return "Aurora MySQL"
}

func (r *RDSClusterInstance) dbInstanceCostComponent(databaseEngine string) *schema.CostComponent {
	purchaseOptionLabel := "on-demand"
	priceFilter := &schema.PriceFilter{
		PurchaseOption: strPtr("on_demand"),
	}

	var err error
	if r.ReservedInstanceTerm != nil {
		resolver := &rdsReservationResolver{
			term:          strVal(r.ReservedInstanceTerm),
			paymentOption: strVal(r.ReservedInstancePaymentOption),
		}
		priceFilter, err = resolver.PriceFilter()
		if err != nil {
			logging.Logger.Warn().Msg(err.Error())
		}
		purchaseOptionLabel = "reserved"
	}

	// Example usage types for Aurora
	// InstanceUsage:db.t3.medium
	// InstanceUsageIOOptimized:db.t3.medium
	// EU-InstanceUsage:db.t3.medium
	// EU-InstanceUsageIOOptimized:db.t3.medium
	usageTypeFilter := "/InstanceUsage:/"
	if r.IOOptimized {
		usageTypeFilter = "/InstanceUsageIOOptimized:/"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Database instance (%s, %s)", purchaseOptionLabel, r.InstanceClass),
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
				{Key: "usagetype", ValueRegex: strPtr(usageTypeFilter)},
			},
		},
		PriceFilter: priceFilter,
	}
}

func (r *RDSClusterInstance) auroraServerlessV2CostComponent(databaseEngine string) *schema.CostComponent {
	var auroraCapacityUnits *decimal.Decimal
	if r.CapacityUnitsPerHr != nil {
		auroraCapacityUnits = decimalPtr(decimal.NewFromFloat(*r.CapacityUnitsPerHr))
	}

	label := "Aurora serverless v2"
	usageType := "Aurora:ServerlessV2Usage$"
	if r.IOOptimized {
		label = "Aurora serverless v2 (I/O-optimized)"
		usageType = "Aurora:ServerlessV2IOOptimizedUsage$"
	}

	return &schema.CostComponent{
		Name:           label,
		Unit:           "ACU-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: auroraCapacityUnits,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("ServerlessV2"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", Value: strPtr(databaseEngine)},
				{Key: "usagetype", ValueRegex: regexPtr(usageType)},
			},
		},
		UsageBased: true,
	}
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
		UsageBased: true,
	}
}
