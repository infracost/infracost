package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
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

	if strings.ToLower(databaseEngineMode) == "serverless" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Aurora serverless",
			Unit:           "ACU-hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: auroraCapacityUnits,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr(databaseEngineMode),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", *databaseEngine))},
				},
			},
		})
	}

	costComponents = append(costComponents, auroraStorageCostComponent(region, u, databaseEngineStorageType)...)

	backupStorageRetention := decimal.NewFromInt(d.Get("backup_retention_period").Int())
	if backupStorageRetention.GreaterThan(decimal.NewFromInt(1)) {

		var snapShotStorageSizeGB, totalBackupStorageGB *decimal.Decimal
		if u != nil && u.Get("backup_snapshot_size_gb").Exists() {
			snapShotStorageSizeGB = decimalPtr(decimal.NewFromFloat(u.Get("backup_snapshot_size_gb").Float()))
			totalBackupStorageGB = decimalPtr(calculateBackupStorage(*snapShotStorageSizeGB, backupStorageRetention))
		}
		costComponents = append(costComponents, auroraBackupStorageCostComponent(region, totalBackupStorageGB, databaseEngine))
	}

	if databaseEngineMode != "Serverless" && !strings.Contains(d.Get("engine").String(), "postgresql") {
		var averageStatements, backtrackChangeRecords, backtrackWindowHours, totalBacktrackChangeRecords *decimal.Decimal
		if u != nil && averageStatementsPerHrExists(u) && changeRecordsPerStatementExists(u) && backtrackWindowHrsExists(u) {
			averageStatements = decimalPtr(decimal.NewFromInt(u.Get("average_statements_per_hr").Int()))
			backtrackChangeRecords = decimalPtr(decimal.NewFromFloat(u.Get("change_records_per_statement").Float()))
			backtrackWindowHours = decimalPtr(decimal.NewFromInt(u.Get("backtrack_window_hrs").Int()))

			totalBacktrackChangeRecords = decimalPtr(calculateBacktrack(*averageStatements, *backtrackChangeRecords, *backtrackWindowHours))
		}
		costComponents = append(costComponents, auroraBacktrackCostComponent(region, totalBacktrackChangeRecords))
	}

	var snapshotExportSizeGB *decimal.Decimal
	if u != nil && u.Get("snapshot_export_size_gb").Exists() {
		snapshotExportSizeGB = decimalPtr(decimal.NewFromFloat(u.Get("snapshot_export_size_gb").Float()))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Snapshot export",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: snapshotExportSizeGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRDS"),
			Region:        strPtr(region),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", *databaseEngine))},
				{Key: "usagetype", ValueRegex: strPtr("/Aurora:SnapshotExportToS3/")},
			},
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func auroraStorageCostComponent(region string, u *schema.UsageData, databaseEngineStorageType *string) []*schema.CostComponent {
	var storageGB, writeRequestsPerSecond, readRequestsPerSecond, monthlyIORequests *decimal.Decimal

	if u != nil && u.Get("storage_gb").Exists() {
		storageGB = decimalPtr(decimal.NewFromFloat(u.Get("storage_gb").Float()))
	}

	if u != nil && u.Get("write_requests_per_sec").Exists() && u.Get("read_requests_per_sec").Exists() {
		writeRequestsPerSecond = decimalPtr(decimal.NewFromInt(u.Get("write_requests_per_sec").Int()))
		readRequestsPerSecond = decimalPtr(decimal.NewFromInt(u.Get("read_requests_per_sec").Int()))
		monthlyIORequests = decimalPtr(calculateIORequests(*readRequestsPerSecond, *writeRequestsPerSecond))
	}

	return []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: storageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", *databaseEngineStorageType))},
					{Key: "usagetype", ValueRegex: strPtr("/Aurora:Storage/")},
				},
			},
		},
		{
			Name:            "I/O rate",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1000000),
			MonthlyQuantity: monthlyIORequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", *databaseEngineStorageType))},
					{Key: "usagetype", ValueRegex: strPtr("/Aurora:Storage/")},
				},
			},
		},
	}
}

func auroraBackupStorageCostComponent(region string, totalBackupStorageGB *decimal.Decimal, databaseEngine *string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: totalBackupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", ValueRegex: strPtr(fmt.Sprintf("/%s/i", *databaseEngine))},
			},
		},
	}
}

func auroraBacktrackCostComponent(region string, backtrackChangeRecords *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backtrack",
		Unit:            "1M change-records",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: backtrackChangeRecords,
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
func averageStatementsPerHrExists(u *schema.UsageData) bool {
	return u.Get("average_statements_per_hr").Exists()
}
func changeRecordsPerStatementExists(u *schema.UsageData) bool {
	return u.Get("change_records_per_statement").Exists()
}
func backtrackWindowHrsExists(u *schema.UsageData) bool {
	return u.Get("backtrack_window_hrs").Exists()
}
