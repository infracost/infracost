package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DocdbCluster struct {
	Address               *string
	Region                *string
	BackupRetentionPeriod *int64
	BackupStorageGb       *float64 `infracost_usage:"backup_storage_gb"`
}

var DocdbClusterUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *DocdbCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocdbCluster) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := []*schema.CostComponent{}

	var retentionPeriod *decimal.Decimal
	if r.BackupRetentionPeriod != nil {
		retentionPeriod = decimalPtr(decimal.NewFromInt(*r.BackupRetentionPeriod))
		if retentionPeriod.GreaterThan(decimal.NewFromInt(1)) {
			var backupStorage *decimal.Decimal
			if r.BackupStorageGb != nil {
				backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGb))
			}
			costComponents = append(costComponents, docDBCluster(region, backupStorage))
		}

	} else {

		var unknown *decimal.Decimal

		costComponents = append(costComponents, docDBCluster(region, unknown))

	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: DocdbClusterUsageSchema,
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
