package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type RDSCluster struct {
	Address                   string
	Region                    string
	EngineMode                string
	Engine                    string
	IOOptimized               bool
	BackupRetentionPeriod     int64
	WriteRequestsPerSec       *int64   `infracost_usage:"write_requests_per_sec"`
	ReadRequestsPerSec        *int64   `infracost_usage:"read_requests_per_sec"`
	ChangeRecordsPerStatement *float64 `infracost_usage:"change_records_per_statement"`
	StorageGB                 *float64 `infracost_usage:"storage_gb"`
	AverageStatementsPerHr    *int64   `infracost_usage:"average_statements_per_hr"`
	BacktrackWindowHrs        *int64   `infracost_usage:"backtrack_window_hrs"`
	SnapshotExportSizeGB      *float64 `infracost_usage:"snapshot_export_size_gb"`
	CapacityUnitsPerHr        *float64 `infracost_usage:"capacity_units_per_hr"`
	BackupSnapshotSizeGB      *float64 `infracost_usage:"backup_snapshot_size_gb"`
}

func (r *RDSCluster) CoreType() string {
	return "RDSCluster"
}

func (r *RDSCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "write_requests_per_sec", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "read_requests_per_sec", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "change_records_per_statement", ValueType: schema.Float64, DefaultValue: 0.0},
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "average_statements_per_hr", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "backtrack_window_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "snapshot_export_size_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "capacity_units_per_hr", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "backup_snapshot_size_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *RDSCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RDSCluster) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	databaseEngineMode := strings.ToLower(r.EngineMode)
	if databaseEngineMode == "" {
		databaseEngineMode = "provisioned"
	}

	databaseEngineStorageType := "Any"

	var databaseEngine string
	switch r.Engine {
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
		databaseEngineStorageType = "(Any|Aurora MySQL)"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
		databaseEngineStorageType = databaseEngine
	}

	var auroraCapacityUnits *decimal.Decimal
	if r.CapacityUnitsPerHr != nil {
		auroraCapacityUnits = decimalPtr(decimal.NewFromFloat(*r.CapacityUnitsPerHr))
	}

	if strings.ToLower(databaseEngineMode) == "serverless" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Aurora serverless",
			Unit:           "ACU-hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: auroraCapacityUnits,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", ValueRegex: regexPtr(databaseEngine)},
				},
			},
			UsageBased: true,
		})
	}

	costComponents = append(costComponents, r.auroraStorageCostComponents(databaseEngineStorageType)...)

	if r.BackupRetentionPeriod > 1 {
		var totalBackupStorageGB *decimal.Decimal

		if r.BackupSnapshotSizeGB != nil {
			totalBackupStorageGB = decimalPtr(r.calculateBackupStorage(decimal.NewFromFloat(*r.BackupSnapshotSizeGB), r.BackupRetentionPeriod))
		}

		costComponents = append(costComponents, r.auroraBackupStorageCostComponent(totalBackupStorageGB, databaseEngine))
	}

	if databaseEngineMode != "serverless" && !strings.Contains(r.Engine, "postgresql") {
		var totalBacktrackChangeRecords *decimal.Decimal

		if r.AverageStatementsPerHr != nil && r.ChangeRecordsPerStatement != nil && r.BacktrackWindowHrs != nil {
			averageStatements := decimal.NewFromInt(*r.AverageStatementsPerHr)
			backtrackChangeRecords := decimal.NewFromFloat(*r.ChangeRecordsPerStatement)
			backtrackWindowHours := decimal.NewFromInt(*r.BacktrackWindowHrs)

			totalBacktrackChangeRecords = decimalPtr(r.calculateBacktrack(averageStatements, backtrackChangeRecords, backtrackWindowHours))
		}

		costComponents = append(costComponents, r.auroraBacktrackCostComponent(totalBacktrackChangeRecords))
	}

	var snapshotExportSizeGB *decimal.Decimal
	if r.SnapshotExportSizeGB != nil {
		snapshotExportSizeGB = decimalPtr(decimal.NewFromFloat(*r.SnapshotExportSizeGB))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Snapshot export",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: snapshotExportSizeGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRDS"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", ValueRegex: regexPtr(databaseEngine)},
				{Key: "usagetype", ValueRegex: regexPtr("Aurora:SnapshotExportToS3$")},
			},
		},
		UsageBased: true,
	})

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *RDSCluster) auroraStorageCostComponents(databaseEngineStorageType string) []*schema.CostComponent {
	var storageGB, writeRequestsPerSecond, readRequestsPerSecond, monthlyIORequests *decimal.Decimal

	if r.StorageGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	if r != nil && r.WriteRequestsPerSec != nil && r.ReadRequestsPerSec != nil {
		writeRequestsPerSecond = decimalPtr(decimal.NewFromInt(*r.WriteRequestsPerSec))
		readRequestsPerSecond = decimalPtr(decimal.NewFromInt(*r.ReadRequestsPerSec))
		monthlyIORequests = decimalPtr(r.calculateIORequests(*readRequestsPerSecond, *writeRequestsPerSecond))
	}

	label := "Storage"
	usageType := "Aurora:StorageUsage$"
	if r.IOOptimized {
		label = "Storage (I/O-optimized)"
		usageType = "Aurora:IO-OptimizedStorageUsage$"
	}

	return []*schema.CostComponent{
		{
			Name:            label,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: storageGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRDS"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "databaseEngine", ValueRegex: regexPtr(databaseEngineStorageType)},
					{Key: "usagetype", ValueRegex: regexPtr(usageType)},
				},
			},
			UsageBased: true,
		},
		{
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
					{Key: "databaseEngine", ValueRegex: regexPtr(databaseEngineStorageType)},
					{Key: "usagetype", ValueRegex: regexPtr("Aurora:StorageIOUsage$")},
				},
			},
			UsageBased: true,
		},
	}
}

func (r *RDSCluster) auroraBackupStorageCostComponent(totalBackupStorageGB *decimal.Decimal, databaseEngine string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: totalBackupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRDS"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "databaseEngine", ValueRegex: regexPtr(databaseEngine)},
				{Key: "usagetype", ValueRegex: regexPtr("Aurora:BackupUsage$")},
			},
		},
		UsageBased: true,
	}
}

func (r *RDSCluster) auroraBacktrackCostComponent(backtrackChangeRecords *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 "Backtrack",
		Unit:                 "1M change-records",
		UnitMultiplier:       decimal.NewFromInt(1000000),
		MonthlyQuantity:      backtrackChangeRecords,
		IgnoreIfMissingPrice: true, // Backtrack is not available in all regions
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRDS"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("Aurora:BacktrackUsage$")},
			},
		},
		UsageBased: true,
	}
}

func (r *RDSCluster) calculateIORequests(writeRequestPerSecond decimal.Decimal, readRequestsPerSecond decimal.Decimal) decimal.Decimal {
	ioPerSecond := writeRequestPerSecond.Add(readRequestsPerSecond)
	monthlyIO := ioPerSecond.Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))
	return monthlyIO
}

func (r *RDSCluster) calculateBackupStorage(snapShotStorageSize decimal.Decimal, numberOfBackups int64) decimal.Decimal {
	return snapShotStorageSize.Mul(decimal.NewFromInt(numberOfBackups)).Sub(snapShotStorageSize)
}

func (r *RDSCluster) calculateBacktrack(averageStatements decimal.Decimal, changeRecords decimal.Decimal, windowHours decimal.Decimal) decimal.Decimal {
	return averageStatements.Mul(decimal.NewFromInt(730)).Mul(changeRecords).Mul(windowHours)
}
