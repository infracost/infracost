package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type FSxWindowsFileSystem struct {
	Address            string
	StorageType        string
	ThroughputCapacity int64
	StorageCapacityGB  int64
	Region             string
	DeploymentType     string
	BackupStorageGB    *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *FSxWindowsFileSystem) CoreType() string {
	return "FSxWindowsFileSystem"
}

func (r *FSxWindowsFileSystem) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *FSxWindowsFileSystem) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *FSxWindowsFileSystem) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.throughputCapacityCostComponent(),
			r.storageCapacityCostComponent(),
			r.backupGBCostComponent(),
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *FSxWindowsFileSystem) deploymentOptionValue() string {
	if strings.Contains(strings.ToLower(r.DeploymentType), "multi_az") {
		return "Multi-AZ"
	}

	return "Single-AZ"
}

func (r *FSxWindowsFileSystem) storageTypeValue() string {
	if strings.ToLower(r.StorageType) == "hdd" {
		return "HDD"
	}

	return "SSD"
}

func (r *FSxWindowsFileSystem) throughputCapacityCostComponent() *schema.CostComponent {
	deploymentOption := r.deploymentOptionValue()

	return &schema.CostComponent{
		Name:            "Throughput capacity",
		Unit:            "MBps",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.ThroughputCapacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Provisioned Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "fileSystemType", Value: strPtr("Windows")},
			},
		},
	}
}

func (r *FSxWindowsFileSystem) storageCapacityCostComponent() *schema.CostComponent {
	deploymentOption := r.deploymentOptionValue()
	storageType := r.storageTypeValue()

	return &schema.CostComponent{
		Name:            fmt.Sprintf("%v storage", storageType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.StorageCapacityGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "fileSystemType", Value: strPtr("Windows")},
				{Key: "storageType", Value: strPtr(storageType)},
			},
		},
	}
}

func (r *FSxWindowsFileSystem) backupGBCostComponent() *schema.CostComponent {
	deploymentOption := r.deploymentOptionValue()

	var backupStorage *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/")},
				{Key: "fileSystemType", Value: strPtr("Windows")},
			},
		},
		UsageBased: true,
	}
}
