package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DocDBClusterSnapshot struct {
	Address         *string
	Region          *string
	BackupStorageGb *int64 `infracost_usage:"backup_storage_gb"`
}

var DocDBClusterSnapshotUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *DocDBClusterSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocDBClusterSnapshot) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := []*schema.CostComponent{}

	var backupStorage *decimal.Decimal
	if r != nil && r.BackupStorageGb != nil {
		backupStorage = decimalPtr(decimal.NewFromInt(*r.BackupStorageGb))
		costComponents = append(costComponents, docDBCluster(region, backupStorage))
	} else {

		var unknown *decimal.Decimal

		costComponents = append(costComponents, docDBCluster(region, unknown))

	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: DocDBClusterSnapshotUsageSchema,
	}
}
