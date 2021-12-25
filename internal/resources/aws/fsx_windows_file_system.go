package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type FSXWindowsFS struct {
	Address            *string
	StorageType        *string
	ThroughputCapacity *int64
	StorageCapacity    *int64
	Region             *string
	DeploymentType     *string
	BackupStorageGb    *int64 `infracost_usage:"backup_storage_gb"`
}

var FSXWindowsFSUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *FSXWindowsFS) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *FSXWindowsFS) BuildResource() *schema.Resource {
	region := *r.Region
	isMultiAZ := strings.Contains(*r.DeploymentType, "MULTI_AZ")
	isHDD := strings.ToLower(*r.StorageType) == "hdd"
	throughput := decimalPtr(decimal.NewFromInt(*r.ThroughputCapacity))
	storageSize := decimalPtr(decimal.NewFromInt(*r.StorageCapacity))

	var backupStorage *decimal.Decimal
	if r != nil && r.BackupStorageGb != nil {
		backupStorage = decimalPtr(decimal.NewFromInt(*r.BackupStorageGb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			throughputCapacity(region, isMultiAZ, throughput),
			storageCapacity(region, isMultiAZ, isHDD, storageSize),
			backupStorageCapacity(region, isMultiAZ, backupStorage),
		}, UsageSchema: FSXWindowsFSUsageSchema,
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
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: throughput,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Provisioned Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "fileSystemType", Value: strPtr("Windows")},
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
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/")},
				{Key: "fileSystemType", Value: strPtr("Windows")},
			},
		},
	}
}
