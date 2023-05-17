package aws

import (
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DBInstance struct {
	Address                                      string
	Region                                       string
	LicenseModel                                 string
	StorageType                                  string
	BackupRetentionPeriod                        int64
	IOOptimized                                  bool
	PerformanceInsightsEnabled                   bool
	PerformanceInsightsLongTermRetention         bool
	MultiAZ                                      bool
	InstanceClass                                string
	Engine                                       string
	IOPS                                         float64
	AllocatedStorageGB                           *float64
	MonthlyStandardIORequests                    *int64   `infracost_usage:"monthly_standard_io_requests"`
	AdditionalBackupStorageGB                    *float64 `infracost_usage:"additional_backup_storage_gb"`
	MonthlyAdditionalPerformanceInsightsRequests *int64   `infracost_usage:"monthly_additional_performance_insights_requests"`
	ReservedInstanceTerm                         *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption                *string  `infracost_usage:"reserved_instance_payment_option"`
}

func (r *DBInstance) CoreType() string {
	return "DBInstance"
}

func (r *DBInstance) UsageSchema() []*schema.UsageItem {
	return DBInstanceUsageSchema
}

var DBInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_standard_io_requests", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "additional_backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_additional_performance_insights_requests", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
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
	if r.MonthlyStandardIORequests != nil {
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
	case "oracle-se", "oracle-se1", "oracle-se2", "oracle-se2-cdb", "oracle-ee", "oracle-ee-cdb":
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
	case "oracle-se2", "oracle-se2-cdb":
		databaseEdition = "Standard Two"
	case "oracle-ee", "oracle-ee-cdb", "sqlserver-ee":
		databaseEdition = "Enterprise"
	case "sqlserver-ex":
		databaseEdition = "Express"
	case "sqlserver-web":
		databaseEdition = "Web"
	}

	var licenseModel string
	engineVal := strings.ToLower(r.Engine)
	if engineVal == "oracle-se1" || engineVal == "oracle-se2" || engineVal == "oracle-se2-cdb" || strings.HasPrefix(engineVal, "sqlserver-") {
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
	iopsDescription := "RDS Provisioned IOPS"

	storageType := strings.ToLower(r.StorageType)
	switch storageType {
	case "io1":
		volumeType = "Provisioned IOPS"
		storageName = "Storage (provisioned IOPS SSD, io1)"
		if iopsVal.LessThan(decimal.NewFromInt(1000)) {
			iopsVal = decimal.NewFromInt(1000)
		}
		if allocatedStorageVal.LessThan(decimal.NewFromInt(100)) {
			allocatedStorageVal = decimal.NewFromInt(100)
		}
	case "standard":
		volumeType = "Magnetic"
		storageName = "Storage (magnetic)"
	case "gp3":
		volumeType = "General Purpose-GP3"
		storageName = "Storage (general purpose SSD, gp3)"
		iopsDescription = "RDS Provisioned GP3 IOPS"

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
	if strings.HasPrefix(databaseEngine, "Aurora") {
		// Example usage types for Aurora
		// InstanceUsage:db.t3.medium
		// InstanceUsageIOOptimized:db.t3.medium
		// EU-InstanceUsage:db.t3.medium
		// EU-InstanceUsageIOOptimized:db.t3.medium
		usageTypeFilter := "/InstanceUsage:/"
		if r.IOOptimized {
			usageTypeFilter = "/InstanceUsageIOOptimized:/"
		}

		instanceAttributeFilters = append(instanceAttributeFilters, &schema.AttributeFilter{
			Key:        "usagetype",
			ValueRegex: strPtr(usageTypeFilter),
		})
	}

	purchaseOptionLabel := "on-demand"
	priceFilter := &schema.PriceFilter{
		PurchaseOption: strPtr("on_demand"),
	}

	var err error
	if r.ReservedInstanceTerm != nil {
		resolver := &rdsReservationResolver{
			term:          strVal(r.ReservedInstanceTerm),
			paymentOption: strVal(r.ReservedInstancePaymentOption),
		}
		priceFilter, err = resolver.PriceFilter()
		if err != nil {
			log.Warnf(err.Error())
		}
		purchaseOptionLabel = "reserved"
	}

	storageFilters := []*schema.AttributeFilter{
		{Key: "deploymentOption", Value: strPtr(deploymentOption)},
		{Key: "databaseEngine", Value: strPtr("Any")},
		{Key: "volumeType", Value: strPtr(volumeType)},
	}

	if storageType == "gp3" {
		storageFilters = append(storageFilters, &schema.AttributeFilter{Key: "usagetype", ValueRegex: strPtr("/\\-RDS\\:GP3\\-Storage$/")})
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database instance (%s, %s, %s)", purchaseOptionLabel, deploymentOption, r.InstanceClass),
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
			PriceFilter: priceFilter,
		},
		{
			Name:            storageName,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &allocatedStorageVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:       strPtr("aws"),
				Region:           strPtr(r.Region),
				Service:          strPtr("AmazonRDS"),
				ProductFamily:    strPtr("Database Storage"),
				AttributeFilters: storageFilters,
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

	if storageType == "io1" || storageType == "gp3" {
		if storageType == "gp3" {
			// For GP3 Storage volumes, all IOPS and throughput use below the baseline is
			// included at no additional charge. For volumes below 400 GiB of allocated
			// storage, the baseline provisioned IOPS is 3,000 and baseline throughput is 125
			// MiBps. Volumes of 400 GiB and above, baseline provisioned IOPS is 12,000 and
			// baseline throughput is 500 MiBps. There is an additional charge for
			// provisioned IOPS and throughput above baseline.
			baseline := decimal.NewFromInt(3000)
			baselineStr := "3,000"
			if allocatedStorageVal.GreaterThanOrEqual(decimal.NewFromInt(400)) {
				baseline = decimal.NewFromInt(12000)
				baselineStr = "12,000"
			}

			if iopsVal.GreaterThan(baseline) {
				over := iopsVal.Sub(baseline)

				costComponents = append(costComponents, &schema.CostComponent{
					Name:            fmt.Sprintf("Provisioned GP3 IOPS (above %s)", baselineStr),
					Unit:            "IOPS",
					UnitMultiplier:  decimal.NewFromInt(1),
					MonthlyQuantity: &over,
					ProductFilter: &schema.ProductFilter{
						VendorName:    strPtr("aws"),
						Region:        strPtr(r.Region),
						Service:       strPtr("AmazonRDS"),
						ProductFamily: strPtr("Provisioned IOPS"),
						AttributeFilters: []*schema.AttributeFilter{
							{Key: "deploymentOption", Value: strPtr(deploymentOption)},
							{Key: "groupDescription", Value: strPtr(iopsDescription)},
							{Key: "databaseEngine", Value: strPtr("Any")},
							{Key: "usagetype", ValueRegex: strPtr("/\\-RDS\\:GP3\\-PIOPS$/")},
						},
					},
				})
			}
		} else {
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
						{Key: "groupDescription", Value: strPtr(iopsDescription)},
						{Key: "databaseEngine", Value: strPtr("Any")},
					},
				},
			})
		}

	}

	var backupStorageGB *decimal.Decimal
	if r.AdditionalBackupStorageGB != nil {
		backupStorageGB = decimalPtr(decimal.NewFromFloat(*r.AdditionalBackupStorageGB))
	}

	if r.BackupRetentionPeriod > 0 || (backupStorageGB != nil && backupStorageGB.GreaterThan(decimal.Zero)) {
		backupStorageDBEngine := "Any"
		attrFilters := []*schema.AttributeFilter{
			{Key: "databaseEngine", Value: strPtr(backupStorageDBEngine)},
			{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/i")},
			{Key: "engineCode", ValueRegex: strPtr("/[0-9]+/")},
			{Key: "operation", Value: strPtr("")},
		}

		if strings.HasPrefix(databaseEngine, "Aurora") {
			backupStorageDBEngine = databaseEngine
			attrFilters = []*schema.AttributeFilter{
				{Key: "databaseEngine", Value: strPtr(backupStorageDBEngine)},
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage/i")},
				{Key: "engineCode", ValueRegex: strPtr("/[0-9]+/")},
			}
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Additional backup storage",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: backupStorageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:       strPtr("aws"),
				Region:           strPtr(r.Region),
				Service:          strPtr("AmazonRDS"),
				ProductFamily:    strPtr("Storage Snapshot"),
				AttributeFilters: attrFilters,
			},
		})
	}

	if r.PerformanceInsightsEnabled {
		if r.PerformanceInsightsLongTermRetention {
			costComponents = append(costComponents, performanceInsightsLongTermRetentionCostComponent(r.Region, r.InstanceClass))
		}

		if r.MonthlyAdditionalPerformanceInsightsRequests == nil || *r.MonthlyAdditionalPerformanceInsightsRequests > 0 {
			costComponents = append(costComponents,
				performanceInsightsAPIRequestCostComponent(r.Region, r.MonthlyAdditionalPerformanceInsightsRequests))
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    DBInstanceUsageSchema,
	}
}

func performanceInsightsLongTermRetentionCostComponent(region, instanceClass string) *schema.CostComponent {
	instanceType := strings.TrimPrefix(instanceClass, "db.")

	vCPUCount := decimal.Zero
	if count, ok := InstanceTypeToVCPU[instanceType]; ok {
		// We were able to lookup thing VCPU count
		vCPUCount = decimal.NewFromInt(count)
	}

	var instanceFamily string
	split := strings.SplitN(instanceType, ".", 2)
	if len(split) > 0 {
		instanceFamily = split[0]
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Performance Insights Long Term Retention (%s)", strings.ToLower(instanceClass)),
		Unit:            "vCPU-month",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &vCPUCount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Performance Insights"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("PI_LTR:" + strings.ToUpper(instanceFamily) + "$")},
			},
		},
	}
}

func performanceInsightsAPIRequestCostComponent(region string, additionalRequests *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Performance Insights API",
		Unit:            "1000 requests",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: intPtrToDecimalPtr(additionalRequests),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Performance Insights"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("PI_API$")},
			},
		},
	}
}
