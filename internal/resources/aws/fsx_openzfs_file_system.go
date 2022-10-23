package aws

import (
	"fmt"
	"math"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type FSxOpenZFSFileSystem struct {
	Address             string
	StorageType         string
	ThroughputCapacity  int64
	ProvisionedIOPS     int64
	ProvisionedIOPSMode string
	StorageCapacityGB   int64
	Region              string
	DeploymentType      string
	DataCompression     string
	CompressionSavings  *int64   `infracost_usage:"compression_savings"`
	BackupStorageGB     *float64 `infracost_usage:"backup_storage_gb"`
}

var FSxOpenZFSFileSystemUsageSchema = []*schema.UsageItem{
	{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *FSxOpenZFSFileSystem) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *FSxOpenZFSFileSystem) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.throughputCapacityCostComponent(),
			r.provisionedIOPSCapacityCostComponent(),
			r.storageCapacityCostComponent(),
			r.backupGBCostComponent(),
		},
		UsageSchema: FSxOpenZFSFileSystemUsageSchema,
	}
}

func (r *FSxOpenZFSFileSystem) throughputCapacityCostComponent() *schema.CostComponent {
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
				{Key: "deploymentOption", Value: strPtr("Single-AZ")},
				{Key: "fileSystemType", Value: strPtr("OpenZFS")},
			},
		},
	}
}

func (r *FSxOpenZFSFileSystem) provisionedIOPSCapacityCostComponent() *schema.CostComponent {
	var provisionedIOPS = decimalPtr(decimal.NewFromInt(0))
	if r.ProvisionedIOPSMode == "USER_PROVISIONED" {
		provisionedIOPS = decimalPtr(decimal.NewFromFloat(math.Max(0, float64(r.ProvisionedIOPS-(3*r.StorageCapacityGB)))))
	}
	return &schema.CostComponent{
		Name:            "Provisioned IOPS",
		Unit:            "IOPS",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: provisionedIOPS,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Provisioned IOPS"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr("Single-AZ")},
				{Key: "fileSystemType", Value: strPtr("OpenZFS")},
			},
		},
	}
}

func (r *FSxOpenZFSFileSystem) storageCapacityCostComponent() *schema.CostComponent {
	var storageCapacity *decimal.Decimal
	var compressionEnabled = ""
	var compressionSavings = int64(0)
	if r.DataCompression != "" && r.DataCompression != "NONE" {
		if r.CompressionSavings != nil {
			compressionSavings = *r.CompressionSavings
		} else {
			compressionSavings = int64(50)
		}
		compressionEnabled = fmt.Sprintf(" (%s compression, %d percent)", r.DataCompression, compressionSavings)
		storageCapacity = decimalPtr(decimal.NewFromInt(int64(math.Round(float64(r.StorageCapacityGB) * float64((1 - float64(compressionSavings)/float64(100)))))))
	} else {
		storageCapacity = decimalPtr(decimal.NewFromInt(r.StorageCapacityGB))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("SSD storage%s", compressionEnabled),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageCapacity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonFSx"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "deploymentOption", Value: strPtr("Single-AZ")},
				{Key: "fileSystemType", Value: strPtr("OpenZFS")},
				{Key: "storageType", Value: strPtr("SSD")},
			},
		},
	}
}

func (r *FSxOpenZFSFileSystem) backupGBCostComponent() *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/")},
				{Key: "fileSystemType", Value: strPtr("OpenZFS")},
			},
		},
	}
}
