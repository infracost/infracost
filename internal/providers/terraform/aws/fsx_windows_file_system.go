package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetFSXWindowsFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_fsx_windows_file_system",
		Notes: []string{"Data deduplication is not supported by Terraform."},
		RFunc: NewFSXWindowsFS,
	}
}

func NewFSXWindowsFS(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	isMultiAZ := strings.Contains(d.Get("deployment_type").String(), "MULTI_AZ")
	isHDD := d.Get("storage_type").String() == "HDD"
	throughput := decimalPtr(decimal.NewFromInt(d.Get("throughput_capacity").Int()))
	storageSize := decimalPtr(decimal.NewFromInt(d.Get("storage_capacity").Int()))

	var backupStorage *decimal.Decimal
	if u != nil && u.Get("backup_storage_gb").Exists() {
		backupStorage = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			throughputCapacity(region, isMultiAZ, throughput),
			storageCapacity(region, isMultiAZ, isHDD, storageSize),
			backupStorageCapacity(region, isMultiAZ, backupStorage),
		},
	}
}

func storageCapacity(region string, isMultiAZ, isHDD bool, storageSize *decimal.Decimal) *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if isMultiAZ {
		deploymentOption = "Multi-AZ"
	}
	storageType := "SDD"
	if isHDD {
		storageType = "HDD"
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%v storage", storageType),
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: storageSize,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "storageType", Value: strPtr(storageType)},
			},
		},
	}
}

func throughputCapacity(region string, isMultiAZ bool, throughput *decimal.Decimal) *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if isMultiAZ {
		deploymentOption = "Multi-AZ"
	}
	return &schema.CostComponent{
		Name:            "Throughput capacity",
		Unit:            "MBps",
		UnitMultiplier:  1,
		MonthlyQuantity: throughput,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Provisioned Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
			},
		},
	}
}

func backupStorageCapacity(region string, isMultiAZ bool, backupStorage *decimal.Decimal) *schema.CostComponent {
	deploymentOption := "Single-AZ"
	if isMultiAZ {
		deploymentOption = "Multi-AZ"
	}
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/")},
			},
		},
	}
}
