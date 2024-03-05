package aws

import (
	"fmt"
	"math"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type FSxOpenZFSFileSystem struct {
	Address                   string
	StorageType               string
	ThroughputCapacity        int64
	ProvisionedIOPS           int64
	ProvisionedIOPSMode       string
	StorageCapacityGB         int64
	Region                    string
	DeploymentType            string
	DataCompression           string
	CompressionSavingsPercent *float64 `infracost_usage:"compression_savings_percent"`
	BackupStorageGB           *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *FSxOpenZFSFileSystem) CoreType() string {
	return "FSxOpenZFSFileSystem"
}

func (r *FSxOpenZFSFileSystem) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
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
		UsageSchema: r.UsageSchema(),
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
	var compressionSavingsPercent float64
	if r.DataCompression != "" && r.DataCompression != "NONE" {
		if r.CompressionSavingsPercent != nil {
			compressionSavingsPercent = *r.CompressionSavingsPercent
		} else {
			compressionSavingsPercent = 50
		}
		compressionEnabled = fmt.Sprintf(" (%s compression, %.0f%%)", r.DataCompression, compressionSavingsPercent)
		storageCapacity = decimalPtr(decimal.NewFromFloat(math.Ceil(float64(r.StorageCapacityGB) * (1 - compressionSavingsPercent/100))))
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

	filters := []*schema.AttributeFilter{
		{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/")},
		{Key: "fileSystemType", Value: strPtr("OpenZFS")},
	}
	if strings.Contains(strings.ToLower(r.DeploymentType), "multi") {
		filters = append(filters, &schema.AttributeFilter{
			Key:   "deploymentOption",
			Value: strPtr("Multi-AZ"),
		})
	} else {
		filters = append(filters, &schema.AttributeFilter{
			Key:   "deploymentOption",
			Value: strPtr("N/A"),
		})
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("aws"),
			Region:           strPtr(r.Region),
			Service:          strPtr("AmazonFSx"),
			ProductFamily:    strPtr("Storage"),
			AttributeFilters: filters,
		},
		UsageBased: true,
	}
}
