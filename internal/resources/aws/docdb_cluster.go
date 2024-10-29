package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DocDBCluster struct {
	Address               string
	Region                string
	BackupRetentionPeriod int64
	BackupStorageGB       *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *DocDBCluster) CoreType() string {
	return "DocDBCluster"
}

func (r *DocDBCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *DocDBCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocDBCluster) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if r.BackupRetentionPeriod > 1 {
		var backupStorage *decimal.Decimal
		if r.BackupStorageGB != nil {
			backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
		}
		costComponents = append(costComponents, r.backupStorageCostComponent(backupStorage))
	} else {
		costComponents = append(costComponents, r.backupStorageCostComponent(nil))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *DocDBCluster) backupStorageCostComponent(backupStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonDocDB"),
			ProductFamily: strPtr("Storage Snapshot"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)BackupUsage$")},
			},
		},
		UsageBased: true,
	}
}
