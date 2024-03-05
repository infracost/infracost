package aws

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type DocDBClusterInstance struct {
	Address             string
	Region              string
	InstanceClass       string
	DataStorageGB       *float64 `infracost_usage:"data_storage_gb"`
	MonthlyIORequests   *int64   `infracost_usage:"monthly_io_requests"`
	MonthlyCPUCreditHrs *int64   `infracost_usage:"monthly_cpu_credit_hrs"`
}

func (r *DocDBClusterInstance) CoreType() string {
	return "DocDBClusterInstance"
}

func (r *DocDBClusterInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "data_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_io_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *DocDBClusterInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocDBClusterInstance) BuildResource() *schema.Resource {
	var storageRate *decimal.Decimal
	if r.DataStorageGB != nil {
		storageRate = decimalPtr(decimal.NewFromFloat(*r.DataStorageGB))
	}

	var ioRequests *decimal.Decimal
	if r.MonthlyIORequests != nil {
		ioRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyIORequests))
	}

	var cpuCredits *decimal.Decimal
	if r.MonthlyCPUCreditHrs != nil {
		cpuCredits = decimalPtr(decimal.NewFromInt(*r.MonthlyCPUCreditHrs))
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (%s, %s)", "on-demand", r.InstanceClass),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Database Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(r.InstanceClass)},
					{Key: "volumeType", Value: strPtr("General Purpose")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name:            "Storage",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: storageRate,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: regexPtr("(^|-)StorageUsage$")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
			UsageBased: true,
		},
		{
			Name:            "I/O requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1000000),
			MonthlyQuantity: ioRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: regexPtr("(^|-)StorageIOUsage$")},
				},
			},
			UsageBased: true,
		},
	}

	if instanceFamily := getBurstableInstanceFamily([]string{"db.t3", "db.t4g"}, r.InstanceClass); instanceFamily != "" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "CPU credits",
			Unit:            "vCPU-hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: cpuCredits,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("CPU Credits"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: regexPtr("CPUCredits:" + instanceFamily + "$")},
				},
			},
			UsageBased: true,
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
