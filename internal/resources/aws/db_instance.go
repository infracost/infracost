package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DBInstance struct {
	Address                   string
	Region                    string
	LicenseModel              string
	StorageType               string
	BackupRetentionPeriod     int64
	MultiAZ                   bool
	InstanceClass             string
	Engine                    string
	IOPS                      float64
	AllocatedStorageGB        *float64
	MonthlyStandardIORequests *int64   `infracost_usage:"monthly_standard_io_requests"`
	AdditionalBackupStorageGB *float64 `infracost_usage:"additional_backup_storage_gb"`
}

var DBInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_standard_io_requests", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "additional_backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *DBInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DBInstance) BuildResource() *schema.Resource {
	deploymentOption := "Single-AZ"
	if r.MultiAZ {
		deploymentOption = "Multi-AZ"
	}

	var monthlyIORequests *decimal.Decimal
	if !(r.MonthlyStandardIORequests == nil) {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(*r.MonthlyStandardIORequests))
	}

	var databaseEngine string
	switch strings.ToLower(r.Engine) {
	case "postgres":
		databaseEngine = "PostgreSQL"
	case "mysql":
		databaseEngine = "MySQL"
	case "mariadb":
		databaseEngine = "MariaDB"
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
	case "oracle-se", "oracle-se1", "oracle-se2", "oracle-ee":
		databaseEngine = "Oracle"
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		databaseEngine = "SQL Server"
	}

	var databaseEdition string
	switch strings.ToLower(r.Engine) {
	case "oracle-se", "sqlserver-se":
		databaseEdition = "Standard"
	case "oracle-se1":
		databaseEdition = "Standard One"
	case "oracle-se2":
		databaseEdition = "Standard Two"
	case "oracle-ee", "sqlserver-ee":
		databaseEdition = "Enterprise"
	case "sqlserver-ex":
		databaseEdition = "Express"
	case "sqlserver-web":
		databaseEdition = "Web"
	}

	var licenseModel string
	engineVal := strings.ToLower(r.Engine)
	if engineVal == "oracle-se1" || engineVal == "oracle-se2" || strings.HasPrefix(engineVal, "sqlserver-") {
		licenseModel = "License included"
	}
	if strings.ToLower(r.LicenseModel) == "bring-your-own-license" {
		licenseModel = "Bring your own license"
	}

	iopsVal := decimal.NewFromFloat(r.IOPS)

	allocatedStorageVal := decimal.NewFromInt(20)
	if r.AllocatedStorageGB != nil {
		allocatedStorageVal = decimal.NewFromFloat(*r.AllocatedStorageGB)
	}

	volumeType := "General Purpose"
	storageName := "Storage (general purpose SSD, gp2)"

	if strings.ToLower(r.StorageType) == "io1" || iopsVal.GreaterThan(decimal.Zero) {
		volumeType = "Provisioned IOPS"
		storageName = "Storage (provisioned IOPS SSD, io1)"
		if iopsVal.LessThan(decimal.NewFromInt(1000)) {
			iopsVal = decimal.NewFromInt(1000)
		}
		if allocatedStorageVal.LessThan(decimal.NewFromInt(100)) {
			allocatedStorageVal = decimal.NewFromInt(100)
		}
	} else if strings.ToLower(r.StorageType) == "standard" {
		volumeType = "Magnetic"
		storageName = "Storage (magnetic)"
	}

	instanceAttributeFilters := []*schema.AttributeFilter{
		{Key: "instanceType", Value: strPtr(r.InstanceClass)},
		{Key: "deploymentOption", Value: strPtr(deploymentOption)},
		{Key: "databaseEngine", Value: strPtr(databaseEngine)},
	}
	if databaseEdition != "" {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "databaseEdition",
			Value: strPtr(databaseEdition),
		})
	}
	if licenseModel != "" {
		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:   "licenseModel",
			Value: strPtr(licenseModel),
		})
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (on-demand, %s, %s)", deploymentOption, r.InstanceClass),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:       strPtr("aws"),
				Region:           strPtr(r.Region),
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
				Region:        strPtr(r.Region),
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
				Region:        strPtr(r.Region),
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
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Provisioned IOPS"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "deploymentOption", Value: strPtr(deploymentOption)},
				},
			},
		})
	}

	var backupStorageGB *decimal.Decimal
	if r.AdditionalBackupStorageGB != nil {
		backupStorageGB = decimalPtr(decimal.NewFromFloat(*r.AdditionalBackupStorageGB))
	}

	if r.BackupRetentionPeriod > 0 || (backupStorageGB != nil && backupStorageGB.GreaterThan(decimal.Zero)) {
		backupStorageDBEngine := "Any"
		if strings.HasPrefix(databaseEngine, "Aurora") {
			backupStorageDBEngine = databaseEngine
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Additional backup storage",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: backupStorageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: strPtr(backupStorageDBEngine)},
					{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/i")},
					{Key: "engineCode", ValueRegex: strPtr("/[0-9]+/")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    DBInstanceUsageSchema,
	}
}
