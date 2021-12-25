package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DBInstance struct {
	Address                   *string
	Region                    *string
	InstanceClass             *string
	Engine                    *string
	LicenseModel              *string
	BackupRetentionPeriod     *string
	MultiAz                   *bool
	Iops                      *float64
	AllocatedStorage          *float64
	StorageType               *string
	MonthlyStandardIoRequests *int64 `infracost_usage:"monthly_standard_io_requests"`
	AdditionalBackupStorageGb *int64 `infracost_usage:"additional_backup_storage_gb"`
}

var DBInstanceUsageSchema = []*schema.UsageItem{{Key: "monthly_standard_io_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "additional_backup_storage_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *DBInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DBInstance) BuildResource() *schema.Resource {
	region := *r.Region

	deploymentOption := "Single-AZ"
	if *r.MultiAz {
		deploymentOption = "Multi-AZ"
	}

	instanceType := *r.InstanceClass

	var monthlyIORequests *decimal.Decimal
	if r != nil && !(r.MonthlyStandardIoRequests == nil) {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(*r.MonthlyStandardIoRequests))
	}

	var databaseEngine *string
	switch strings.ToLower(*r.Engine) {
	case "postgres":
		databaseEngine = strPtr("PostgreSQL")
	case "mysql":
		databaseEngine = strPtr("MySQL")
	case "mariadb":
		databaseEngine = strPtr("MariaDB")
	case "aurora", "aurora-mysql":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
	case "oracle-se", "oracle-se1", "oracle-se2", "oracle-ee":
		databaseEngine = strPtr("Oracle")
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		databaseEngine = strPtr("SQL Server")
	}

	var databaseEdition *string
	switch strings.ToLower(*r.Engine) {
	case "oracle-se", "sqlserver-se":
		databaseEdition = strPtr("Standard")
	case "oracle-se1":
		databaseEdition = strPtr("Standard One")
	case "oracle-se2":
		databaseEdition = strPtr("Standard Two")
	case "oracle-ee", "sqlserver-ee":
		databaseEdition = strPtr("Enterprise")
	case "sqlserver-ex":
		databaseEdition = strPtr("Express")
	case "sqlserver-web":
		databaseEdition = strPtr("Web")
	}

	var licenseModel *string
	engineVal := strings.ToLower(*r.Engine)
	if engineVal == "oracle-se1" || engineVal == "oracle-se2" || strings.HasPrefix(engineVal, "sqlserver-") {
		licenseModel = strPtr("License included")
	}
	if strings.ToLower(*r.LicenseModel) == "bring-your-own-license" {
		licenseModel = strPtr("Bring your own license")
	}

	iopsVal := decimal.Zero
	if !(r.Iops == nil) {
		iopsVal = decimal.NewFromFloat(*r.Iops)
	}

	allocatedStorageVal := decimal.NewFromInt(20)
	if !(r.AllocatedStorage == nil) {
		allocatedStorageVal = decimal.NewFromFloat(*r.AllocatedStorage)
	}

	volumeType := "General Purpose"
	storageName := "Storage (general purpose SSD, gp2)"
	if !(r.StorageType == nil) {
		if strings.ToLower(*r.StorageType) == "io1" || iopsVal.GreaterThan(decimal.Zero) {
			volumeType = "Provisioned IOPS"
			storageName = "Storage (provisioned IOPS SSD, io1)"
			if iopsVal.LessThan(decimal.NewFromInt(1000)) {
				iopsVal = decimal.NewFromInt(1000)
			}
			if allocatedStorageVal.LessThan(decimal.NewFromInt(100)) {
				allocatedStorageVal = decimal.NewFromInt(100)
			}
		} else if strings.ToLower(*r.StorageType) == "standard" {
			volumeType = "Magnetic"
			storageName = "Storage (magnetic)"
		}
	}

	instanceAttributeFilters := []*schema.AttributeFilter{
		{Key: "instanceType", Value: strPtr(instanceType)},
		{Key: "deploymentOption", Value: strPtr(deploymentOption)},
		{Key: "databaseEngine", Value: databaseEngine},
	}
	if databaseEdition != nil {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "databaseEdition",
			Value: databaseEdition,
		})
	}
	if licenseModel != nil {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "licenseModel",
			Value: licenseModel,
		})
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (on-demand, %s, %s)", deploymentOption, instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:       strPtr("aws"),
				Region:           strPtr(region),
				Service:          strPtr("AmazonRDS"),
				ProductFamily:    strPtr("Database Instance"),
				AttributeFilters: instanceAttributeFilters,
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name:            storageName,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &allocatedStorageVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeType", Value: strPtr(volumeType)},
					{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				},
			},
		},
	}

	if strings.ToLower(volumeType) == "magnetic" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "I/O requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1000000),
			MonthlyQuantity: monthlyIORequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/RDS:StorageIOUsage/i")},
				},
			},
		})
	}

	if strings.ToLower(volumeType) == "provisioned iops" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Provisioned IOPS",
			Unit:            "IOPS",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &iopsVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Provisioned IOPS"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				},
			},
		})
	}

	if !(r.BackupRetentionPeriod == nil) || (r != nil && !(r.AdditionalBackupStorageGb == nil)) {
		var backupStorageGB *decimal.Decimal
		if r != nil && !(r.AdditionalBackupStorageGb == nil) {
			backupStorageGB = decimalPtr(decimal.NewFromInt(*r.AdditionalBackupStorageGb))
		}

		backupStorageDbEngine := "Any"
		if databaseEngine != nil && strings.HasPrefix(*databaseEngine, "Aurora") {
			backupStorageDbEngine = *databaseEngine
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Additional backup storage",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: backupStorageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: strPtr(backupStorageDbEngine)},
					{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/i")},
					{Key: "engineCode", ValueRegex: strPtr("/[0-9]+/")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: DBInstanceUsageSchema,
	}
}
