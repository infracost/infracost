package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type RDSCluster struct {
	Address                   *string
	Region                    *string
	EngineMode                *string
	Engine                    *string
	BackupRetentionPeriod     *int64
	ReadRequestsPerSec        *int64   `infracost_usage:"read_requests_per_sec"`
	CapacityUnitsPerHr        *int64   `infracost_usage:"capacity_units_per_hr"`
	ChangeRecordsPerStatement *float64 `infracost_usage:"change_records_per_statement"`
	BacktrackWindowHrs        *int64   `infracost_usage:"backtrack_window_hrs"`
	SnapshotExportSizeGb      *float64 `infracost_usage:"snapshot_export_size_gb"`
	BackupSnapshotSizeGb      *float64 `infracost_usage:"backup_snapshot_size_gb"`
	AverageStatementsPerHr    *int64   `infracost_usage:"average_statements_per_hr"`
	StorageGb                 *float64 `infracost_usage:"storage_gb"`
	WriteRequestsPerSec       *int64   `infracost_usage:"write_requests_per_sec"`
}

var RDSClusterUsageSchema = []*schema.UsageItem{{Key: "read_requests_per_sec", ValueType: schema.Int64, DefaultValue: 0}, {Key: "capacity_units_per_hr", ValueType: schema.Int64, DefaultValue: 0}, {Key: "change_records_per_statement", ValueType: schema.Float64, DefaultValue: 0.000000}, {Key: "backtrack_window_hrs", ValueType: schema.Int64, DefaultValue: 0}, {Key: "snapshot_export_size_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "backup_snapshot_size_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "average_statements_per_hr", ValueType: schema.Int64, DefaultValue: 0}, {Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "write_requests_per_sec", ValueType: schema.Int64, DefaultValue: 0}}

func (r *RDSCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RDSCluster) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := make([]*schema.CostComponent, 0)

	databaseEngineMode := "provisioned"
	if r.EngineMode != nil {
		databaseEngineMode = strings.Title(*r.EngineMode)
	}

	databaseEngineStorageType := strPtr("Any")

	var databaseEngine *string
	switch *r.Engine {
	case "aurora", "aurora-mysql":
		databaseEngine = strPtr("Aurora MySQL")
	case "aurora-postgresql":
		databaseEngine = strPtr("Aurora PostgreSQL")
		databaseEngineStorageType = databaseEngine
	}

	var auroraCapacityUnits *decimal.Decimal
	if r != nil && r.CapacityUnitsPerHr != nil {
		auroraCapacityUnits = decimalPtr(decimal.NewFromInt(*r.CapacityUnitsPerHr))
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

	costComponents = append(costComponents, auroraStorageCostComponent(region, r, databaseEngineStorageType)...)

	backupStorageRetention := decimal.NewFromInt(*r.BackupRetentionPeriod)
	if backupStorageRetention.GreaterThan(decimal.NewFromInt(1)) {

		var snapShotStorageSizeGB, totalBackupStorageGB *decimal.Decimal
		if r != nil && r.BackupSnapshotSizeGb != nil {
			snapShotStorageSizeGB = decimalPtr(decimal.NewFromFloat(*r.BackupSnapshotSizeGb))
			totalBackupStorageGB = decimalPtr(calculateBackupStorage(*snapShotStorageSizeGB, backupStorageRetention))
		}
		costComponents = append(costComponents, auroraBackupStorageCostComponent(region, totalBackupStorageGB, databaseEngine))
	}

	if databaseEngineMode != "Serverless" && !strings.Contains(*r.Engine, "postgresql") {
		var averageStatements, backtrackChangeRecords, backtrackWindowHours, totalBacktrackChangeRecords *decimal.Decimal
		if r != nil && averageStatementsPerHrExists(r) && changeRecordsPerStatementExists(r) && backtrackWindowHrsExists(r) {
			averageStatements = decimalPtr(decimal.NewFromInt(*r.AverageStatementsPerHr))
			backtrackChangeRecords = decimalPtr(decimal.NewFromFloat(*r.ChangeRecordsPerStatement))
			backtrackWindowHours = decimalPtr(decimal.NewFromInt(*r.BacktrackWindowHrs))

			totalBacktrackChangeRecords = decimalPtr(calculateBacktrack(*averageStatements, *backtrackChangeRecords, *backtrackWindowHours))
		}
		costComponents = append(costComponents, auroraBacktrackCostComponent(region, totalBacktrackChangeRecords))
	}

	var snapshotExportSizeGB *decimal.Decimal
	if r != nil && r.SnapshotExportSizeGb != nil {
		snapshotExportSizeGB = decimalPtr(decimal.NewFromFloat(*r.SnapshotExportSizeGb))
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
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: RDSClusterUsageSchema,
	}
}

func auroraStorageCostComponent(region string, r *RDSCluster, databaseEngineStorageType *string) []*schema.CostComponent {
	var storageGB, writeRequestsPerSecond, readRequestsPerSecond, monthlyIORequests *decimal.Decimal

	if r != nil && r.StorageGb != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGb))
	}

	if r != nil && r.WriteRequestsPerSec != nil && r.ReadRequestsPerSec != nil {
		writeRequestsPerSecond = decimalPtr(decimal.NewFromInt(*r.WriteRequestsPerSec))
		readRequestsPerSecond = decimalPtr(decimal.NewFromInt(*r.ReadRequestsPerSec))
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
func averageStatementsPerHrExists(r *RDSCluster,) bool {
	return r.AverageStatementsPerHr != nil
}
func changeRecordsPerStatementExists(r *RDSCluster,) bool {
	return r.ChangeRecordsPerStatement != nil
}
func backtrackWindowHrsExists(r *RDSCluster,) bool {
	return r.BacktrackWindowHrs != nil
}
