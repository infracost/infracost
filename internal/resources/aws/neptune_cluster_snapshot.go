package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type NeptuneClusterSnapshot struct {
	Address               *string
	Region                *string
	BackupRetentionPeriod *int64
	BackupStorageGb       *float64 `infracost_usage:"backup_storage_gb"`
}

var NeptuneClusterSnapshotUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *NeptuneClusterSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneClusterSnapshot) BuildResource() *schema.Resource {
	if r.BackupRetentionPeriod != nil && *r.BackupRetentionPeriod < 2 {
		return &schema.Resource{
			Name:        *r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: NeptuneClusterSnapshotUsageSchema,
		}
	}

	cluster := &NeptuneCluster{
		Address:         r.Address,
		Region:          r.Region,
		BackupStorageGb: r.BackupStorageGb,
	}
	backupCostComponent := cluster.backupCostComponent()

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: []*schema.CostComponent{backupCostComponent}, UsageSchema: NeptuneClusterSnapshotUsageSchema,
	}
}
