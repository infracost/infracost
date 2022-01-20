package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DocDBClusterInstance struct {
	Address             string
	Region              string
	InstanceClass       string
	DataStorageGB       *float64 `infracost_usage:"data_storage_gb"`
	MonthlyIORequests   *int64   `infracost_usage:"monthly_io_requests"`
	MonthlyCPUCreditHrs *int64   `infracost_usage:"monthly_cpu_credit_hrs"`
}

var DocDBClusterInstanceUsageSchema = []*schema.UsageItem{
	{Key: "data_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_io_requests", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "monthly_cpu_credit_hrs", ValueType: schema.Int64, DefaultValue: 0},
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

	var cpuCreditsT3 *decimal.Decimal
	if r.MonthlyCPUCreditHrs != nil {
		cpuCreditsT3 = decimalPtr(decimal.NewFromInt(*r.MonthlyCPUCreditHrs))
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
					{Key: "usagetype", Value: strPtr("StorageUsage")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
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
					{Key: "usagetype", Value: strPtr("StorageIOUsage")},
				},
			},
		},
	}

	if strings.HasPrefix(r.InstanceClass, "db.t3.") {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "CPU credits",
			Unit:            "vCPU-hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: cpuCreditsT3,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("CPU Credits"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("CPUCredits:db.t3")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    DocDBClusterInstanceUsageSchema,
	}
}
