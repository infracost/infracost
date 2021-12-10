package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetNeptuneClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster",
		RFunc: NewNeptuneCluster,
	}
}

func NewNeptuneCluster(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var storageGb, monthlyIoRequests *decimal.Decimal
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("storage_gb").Type != gjson.Null {
		storageGb = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
	}
	if u != nil && u.Get("monthly_io_requests").Type != gjson.Null {
		monthlyIoRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_io_requests").Int()))
	}

	costComponents = append(costComponents, neptuneClusterStorageIOsCostComponents("Storage", "GB", region, "StorageUsage", storageGb, 1))
	costComponents = append(costComponents, neptuneClusterStorageIOsCostComponents("I/O requests", "1M request", region, "StorageIOUsage", monthlyIoRequests, 1000000))
	var retentionPeriod *decimal.Decimal
	if d.Get("backup_retention_period").Type != gjson.Null {
		retentionPeriod = decimalPtr(decimal.NewFromInt(d.Get("backup_retention_period").Int()))
		if retentionPeriod.GreaterThan(decimal.NewFromInt(1)) {
			costComponents = append(costComponents, backupCostComponent(u, region))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func backupCostComponent(u *schema.UsageData, region string) *schema.CostComponent {

	var backupStorage *decimal.Decimal
	if u != nil && u.Get("backup_storage_gb").Type != gjson.Null {
		backupStorage = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
	}

	return neptuneClusterBackupCostComponent(region, backupStorage)
}
func neptuneClusterStorageIOsCostComponents(name, unit, region, usageType string, quantity *decimal.Decimal, unitMulti int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(int64(unitMulti)),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

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
