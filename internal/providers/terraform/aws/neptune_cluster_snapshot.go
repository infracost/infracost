package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetNeptuneClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster_snapshot",
		RFunc: NewNeptuneClusterSnapshot,
		ReferenceAttributes: []string{
			"db_cluster_identifier",
		},
	}
}

func NewNeptuneClusterSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var resourceData *schema.ResourceData
	dbClusterIdentifier := d.References("db_cluster_identifier")
	var retentionPeriod *decimal.Decimal

	if len(dbClusterIdentifier) > 0 {
		resourceData = dbClusterIdentifier[0]
		if resourceData.Get("backup_retention_period").Type != gjson.Null {
			retentionPeriod = decimalPtr(decimal.NewFromInt(resourceData.Get("backup_retention_period").Int()))
			if retentionPeriod.LessThan(decimal.NewFromInt(2)) {
				return &schema.Resource{
					Name:      d.Address,
					NoPrice:   true,
					IsSkipped: true,
				}
			}
		}
	}
	region := d.Get("region").String()

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: []*schema.CostComponent{backupCostComponent(u, region)},
	}
}

// Fix for migrated missing funcs, remove after migration
func backupCostComponent(u *schema.UsageData, region string) *schema.CostComponent {

	var backupStorage *decimal.Decimal
	if u != nil && u.Get("backup_storage_gb").Type != gjson.Null {
		backupStorage = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
	}

	return neptuneClusterBackupCostComponent(region, backupStorage)
}

// Fix for migrated missing funcs, remove after migration
func neptuneClusterBackupCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
