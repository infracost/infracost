package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDocDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster",
		RFunc: NewDocDBCluster,
	}

}

func NewDocDBCluster(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{}

	var retentionPeriod *decimal.Decimal
	if d.Get("backup_retention_period").Exists() {
		retentionPeriod = decimalPtr(decimal.NewFromInt(d.Get("backup_retention_period").Int()))
		if retentionPeriod.GreaterThan(decimal.NewFromInt(1)) {
			var backupStorage *decimal.Decimal
			if u != nil && u.Get("backup_storage_gb").Exists() {
				backupStorage = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
			}
			costComponents = append(costComponents, docDBCluster(region, backupStorage))
		}

	} else {

		var unknown *decimal.Decimal

		costComponents = append(costComponents, docDBCluster(region, unknown))

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func docDBCluster(region string, backupStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonDocDB"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("BackupUsage")},
			},
		},
	}
}
