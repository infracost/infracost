package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"strings"
)

func GetRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster",
		RFunc: NewRDSCluster,
	}
}

func NewRDSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	costComponents := make([]*schema.CostComponent, 0)

	databaseEngineMode := "provisioned"
	if d.Get("engine_mode").Exists() {
		databaseEngineMode = strings.Title(d.Get("engine_mode").String())
	}

	databaseEngineStorageType := strPtr("Any")

	var databaseEngine *string
	switch d.Get("engine").String() {
	case "aurora", "aurora-mysql":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
		databaseEngineStorageType = databaseEngine
	}

	var auroraCapacityUnits *decimal.Decimal
	if u != nil && u.Get("capacity_unit_hrs").Exists() {
		auroraCapacityUnits = decimalPtr(decimal.NewFromInt(u.Get("capacity_unit_hrs").Int()))
	}

	if databaseEngineMode == "Serverless" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Aurora serverless",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: auroraCapacityUnits,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr(databaseEngineMode),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: databaseEngine},
				},
			},
		})
	}

	costComponents = append(costComponents, auroraStorageCostComponent(region, u, databaseEngineStorageType)...)

	backupStorageRetention := decimal.NewFromInt(d.Get("backup_retention_period").Int())
	if backupStorageRetention.GreaterThan(decimal.NewFromInt(1)) {
		snapShotStorageSizeGB := decimal.Zero
		if u != nil && u.Get("backup_snapshot_size_gb").Exists() {
			snapShotStorageSizeGB = decimal.NewFromFloat(u.Get("backup_snapshot_size_gb").Float())
		}

		totalBackupStorageGB := calculateBackupStorage(snapShotStorageSizeGB, backupStorageRetention)

		if totalBackupStorageGB.GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, auroraBackupStorageCostComponent(region, totalBackupStorageGB, databaseEngineStorageType))
		}
	}

	backtrackChangeRecords := decimal.Zero
	if u != nil && u.Get("backtrack_change_records").Exists() {
		backtrackChangeRecords = decimal.NewFromInt(u.Get("backtrack_change_records").Int())
	}

	backtrackPeriod := decimal.Zero
	if u != nil && u.Get("backtrack_change_record_hrs").Exists() {
		backtrackPeriod = decimal.NewFromInt(u.Get("backtrack_change_record_hrs").Int())
	}

	totalBacktrackChangeRecords := backtrackChangeRecords.Mul(backtrackPeriod)
	if totalBacktrackChangeRecords.GreaterThan(decimal.Zero) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Backtrack",
			Unit:            "Record-hours",
			UnitMultiplier:  1000000,
			MonthlyQuantity: decimalPtr(totalBacktrackChangeRecords),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("Aurora:BacktrackUsage")},
				},
			},
		})
	}

	snapshotExportSizeGB := decimal.Zero
	if u != nil && u.Get("snapshot_export_size_gb").Exists() {
		snapshotExportSizeGB = decimal.NewFromFloat(u.Get("snapshot_export_size_gb").Float())
	}

	if snapshotExportSizeGB.GreaterThan(decimal.Zero) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Snapshot export",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(snapshotExportSizeGB),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: databaseEngine},
					{Key: "usagetype", Value: strPtr("Aurora:SnapshotExportToS3")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func auroraStorageCostComponent(region string, u *schema.UsageData, databaseEngineStorageType *string) []*schema.CostComponent {
	storageGB := decimal.Zero
	if u != nil && u.Get("storage_gb").Exists() {
		storageGB = decimal.NewFromFloat(u.Get("storage_gb").Float())
	}

	writeRequestsPerSecond := decimal.Zero
	if u != nil && u.Get("read_request_secs").Exists() {
		writeRequestsPerSecond = decimal.NewFromInt(u.Get("write_request_secs").Int())
	}

	readRequestsPerSecond := decimal.Zero
	if u != nil && u.Get("read_request_secs").Exists() {
		readRequestsPerSecond = decimal.NewFromInt(u.Get("read_request_secs").Int())
	}

	monthlyIORequests := calculateIORequests(readRequestsPerSecond, writeRequestsPerSecond)

	return []*schema.CostComponent{
		{
			Name:            "Storage rate",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: &storageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: databaseEngineStorageType},
					{Key: "usagetype", ValueRegex: strPtr("/Aurora:Storage/")},
				},
			},
		},
		{
			Name:            "I/O rate",
			Unit:            "requests",
			UnitMultiplier:  1000000,
			MonthlyQuantity: &monthlyIORequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: databaseEngineStorageType},
					{Key: "usagetype", ValueRegex: strPtr("/Aurora:Storage/")},
				},
			},
		},
	}
}

func auroraBackupStorageCostComponent(region string, totalBackupStorageGB decimal.Decimal, databaseEngine *string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: &totalBackupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", Value: databaseEngine},
			},
		},
	}
}

func calculateIORequests(writeRequestPerSecond decimal.Decimal, readRequestsPerSecond decimal.Decimal) decimal.Decimal {
	ioPerSecond := writeRequestPerSecond.Add(readRequestsPerSecond)
	monthlyIO := ioPerSecond.Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))
	return monthlyIO
}

func calculateBackupStorage(snapShotStorageSize decimal.Decimal, numberOfBackups decimal.Decimal) decimal.Decimal {
	return snapShotStorageSize.Mul(numberOfBackups).Sub(snapShotStorageSize)
}
