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
	if u != nil && u.Get("capacity_units_per_hr").Exists() {
		auroraCapacityUnits = decimalPtr(decimal.NewFromInt(u.Get("capacity_units_per_hr").Int()))
	}

	if databaseEngineMode == "Serverless" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Aurora serverless",
			Unit:           "ACU-hours",
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
			costComponents = append(costComponents, auroraBackupStorageCostComponent(region, totalBackupStorageGB, databaseEngine))
		}
	}

	if databaseEngineMode != "Serverless" && !strings.Contains(d.Get("engine").String(), "postgresql") {
		averageStatements := decimal.Zero
		if u != nil && u.Get("average_statements_per_hr").Exists() {
			averageStatements = decimal.NewFromInt(u.Get("average_statements_per_hr").Int())
		}

		backtrackChangeRecords := decimal.Zero
		if u != nil && u.Get("change_records_per_statement").Exists() {
			backtrackChangeRecords = decimal.NewFromFloat(u.Get("change_records_per_statement").Float())
		}

		backtrackWindowHours := decimal.Zero
		if u != nil && u.Get("backtrack_window_hrs").Exists() {
			backtrackWindowHours = decimal.NewFromInt(u.Get("backtrack_window_hrs").Int())
		}

		totalBacktrackChangeRecords := calculateBacktrack(averageStatements, backtrackChangeRecords, backtrackWindowHours)
		if totalBacktrackChangeRecords.GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, auroraBacktrackCostComponent(region, totalBacktrackChangeRecords))
		}
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
				Region:        strPtr(region),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", Value: databaseEngine},
					{Key: "usagetype", ValueRegex: strPtr("/Aurora:SnapshotExportToS3/")},
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
	if u != nil && u.Get("write_requests_per_sec").Exists() {
		writeRequestsPerSecond = decimal.NewFromInt(u.Get("write_requests_per_sec").Int())
	}

	readRequestsPerSecond := decimal.Zero
	if u != nil && u.Get("read_requests_per_sec").Exists() {
		readRequestsPerSecond = decimal.NewFromInt(u.Get("read_requests_per_sec").Int())
	}

	monthlyIORequests := calculateIORequests(readRequestsPerSecond, writeRequestsPerSecond)

	return []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB",
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
			Unit:            "1M requests",
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
		Unit:            "GB",
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

func auroraBacktrackCostComponent(region string, backtrackChangeRecords decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backtrack",
		Unit:            "1M change-records",
		UnitMultiplier:  1000000,
		MonthlyQuantity: decimalPtr(backtrackChangeRecords),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRDS"),
			Region:        strPtr(region),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Aurora:BacktrackUsage/")},
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

func calculateBacktrack(averageStatements decimal.Decimal, changeRecords decimal.Decimal, windowHours decimal.Decimal) decimal.Decimal {
	return averageStatements.Mul(decimal.NewFromInt(730)).Mul(changeRecords).Mul(windowHours)
}
