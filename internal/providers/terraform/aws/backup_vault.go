package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetBackupVaultRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_backup_vault",
		RFunc: NewBackupVault,
	}
}

type backupData struct {
	ref       string
	name      string
	unit      string
	usageType string
	service   string
	family    string
	key       string
	value     string
	qty       *decimal.Decimal
}

func NewBackupVault(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := []*schema.CostComponent{}

	data := []backupData{
		{ref: "monthly_efs_warm_backup_gb", name: "EFS backup (warm)", unit: "GB", usageType: "WarmStorage-ByteHrs-EFS", service: "AWSBackup", family: "AWS Backup Storage"},
		{ref: "monthly_efs_cold_backup_gb", name: "EFS backup (cold)", unit: "GB", usageType: "EarlyDelete-ColdByteHrs-EFS", service: "AWSBackup", family: "AWS Backup Early Delete Size"},
		{ref: "monthly_efs_warm_restore_gb", name: "EFS restore (warm)", unit: "GB", usageType: "PartialRestore-Warm-EFS", service: "AWSBackup", family: "AWS Backup Storage"},
		{ref: "monthly_efs_cold_restore_gb", name: "EFS restore (cold)", unit: "GB", usageType: "PartialRestore-Cold-EFS", service: "AWSBackup", family: "AWS Backup Storage"},
		{ref: "monthly_efs_item_restore_requests", name: "EFS restore (item-level)", unit: "requests", usageType: "PartialRestore-Jobs-EFS", service: "AWSBackup", family: "AWS Backup Storage"},
		{ref: "monthly_ebs_snapshot_gb", name: "EBS snapshot", unit: "GB", usageType: "EBS:SnapshotUsage$", service: "AmazonEC2", family: "Storage Snapshot"},
		{ref: "monthly_rds_snapshot_gb", name: "RDS snapshot", unit: "GB", usageType: "RDS:ChargedBackupUsage", service: "AmazonRDS", family: "Storage Snapshot"},
		{ref: "monthly_dynamodb_backup_gb", name: "DynamoDB backup", unit: "GB", usageType: "TimedBackupStorage-ByteHrs", service: "AmazonDynamoDB", family: "Amazon DynamoDB On-Demand Backup Storage"},
		{ref: "monthly_dynamodb_restore_gb", name: "DynamoDB restore", unit: "GB", usageType: "RestoreDataSize-Bytes", service: "AmazonDynamoDB", family: "Amazon DynamoDB Restore Data Size"},
		//	{ref: "monthly_storage_gateway_backup_gb", name: "Storage gateway backup", unit: "GB", usageType: "", service: "", family: ""},
	}

	for _, d := range data {
		if u != nil && u.Get(d.ref).Type != gjson.Null {
			d.qty = decimalPtr(decimal.NewFromInt(u.Get(d.ref).Int()))
		}

		costComponents = append(costComponents, backupVaultCostComponent(region, d))
	}

	additData := []backupData{
		{ref: "monthly_aurora_snapshot_gb", name: "Aurora snapshot", unit: "GB",
			usageType: "Aurora:BackupUsage", service: "AmazonRDS", family: "Storage Snapshot", key: "databaseEngine", value: "Aurora PostgreSQL"}, // Prices of PostgreSQL and MySQL are equals for all regions.

		{ref: "monthly_fsx_windows_backup_gb", name: "FSx for windows backup", unit: "GB", usageType: "BackupUsage", service: "AmazonFSx", family: "Storage",
			key: "fileSystemType", value: "Lustre"},

		{ref: "monthly_fsx_lustre_backup_gb", name: "FSx for lustre backup", unit: "GB", usageType: "BackupUsage", service: "AmazonFSx", family: "Storage",
			key: "fileSystemType", value: "Lustre"}, // Prices for Windows/Lustre backup are equals for all regions.
	}

	for _, d := range additData {
		if u != nil && u.Get(d.ref).Type != gjson.Null {
			d.qty = decimalPtr(decimal.NewFromInt(u.Get(d.ref).Int()))
		}

		costComponents = append(costComponents, additionalBackupVaultCostComponent(region, d))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func backupVaultCostComponent(region string, d backupData) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            d.name,
		Unit:            d.unit,
		UnitMultiplier:  1,
		MonthlyQuantity: d.qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr(d.service),
			ProductFamily: strPtr(d.family),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", d.usageType))},
			},
		},
	}
}

func additionalBackupVaultCostComponent(region string, d backupData) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            d.name,
		Unit:            d.unit,
		UnitMultiplier:  1,
		MonthlyQuantity: d.qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr(d.service),
			ProductFamily: strPtr(d.family),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", d.usageType))},
				{Key: d.key, Value: strPtr(d.value)},
			},
		},
	}
}
