package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDocDBClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster_instance",
		RFunc: NewDocDBClusterInstance,
	}
}

func NewDocDBClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	instanceType := d.Get("instance_class").String()

	var storageRate *decimal.Decimal
	if u != nil && u.Get("monthly_data_storage_gb").Exists() {
		storageRate = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_storage_gb").Int()))
	}

	var ioRequests *decimal.Decimal
	if u != nil && u.Get("monthly_input_output_operations").Exists() {
		ioRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_input_output_operations").Int()))
	}

	var backupStorage *decimal.Decimal
	if u != nil && u.Get("monthly_backup_storage_gb").Exists() {
		backupStorage = decimalPtr(decimal.NewFromInt(u.Get("monthly_backup_storage_gb").Int()))
	}

	var cpuCreditsT3 *decimal.Decimal
	if u != nil && u.Get("cpu_credits").Exists() {
		cpuCreditsT3 = decimalPtr(decimal.NewFromInt(u.Get("cpu_credits").Int()))
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (%s, %s)", "on-demand", instanceType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Database Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name:            "Storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: storageRate,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
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
			Name:            "I/O",
			Unit:            "requests",
			UnitMultiplier:  1000000,
			MonthlyQuantity: ioRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("StorageIOUsage")},
				},
			},
		},
		{
			Name:            "Backup storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: backupStorage,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("BackupUsage")},
				},
			},
		},
	}

	if strings.HasPrefix(instanceType, "db.t3.") {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "CPU credits",
			Unit:            "vCPU-hours",
			UnitMultiplier:  1,
			MonthlyQuantity: cpuCreditsT3,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("CPU Credits"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("CPUCredits:db.t3")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
