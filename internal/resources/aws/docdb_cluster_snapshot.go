package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DocdbClusterSnapshot struct {
	Address         *string
	Region          *string
	BackupStorageGb *float64 `infracost_usage:"backup_storage_gb"`
}

var DocdbClusterSnapshotUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *DocdbClusterSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocdbClusterSnapshot) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := []*schema.CostComponent{}

	var backupStorage *decimal.Decimal
	if r.BackupStorageGb != nil {
		backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGb))
		costComponents = append(costComponents, docDBCluster(region, backupStorage))
	} else {

		var unknown *decimal.Decimal

		costComponents = append(costComponents, docDBCluster(region, unknown))

	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: DocdbClusterSnapshotUsageSchema,
	}
}
