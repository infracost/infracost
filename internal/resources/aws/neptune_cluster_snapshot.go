package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type NeptuneClusterSnapshot struct {
	Address               string
	Region                string
	BackupRetentionPeriod *int64   // This can be unknown since it's retrieved from the Neptune cluster
	BackupStorageGB       *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *NeptuneClusterSnapshot) CoreType() string {
	return "NeptuneClusterSnapshot"
}

func (r *NeptuneClusterSnapshot) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *NeptuneClusterSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneClusterSnapshot) BuildResource() *schema.Resource {
	if r.BackupRetentionPeriod != nil && *r.BackupRetentionPeriod < 2 {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	cluster := &NeptuneCluster{
		Region:          r.Region,
		BackupStorageGB: r.BackupStorageGB,
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{cluster.backupStorageCostComponent()},
		UsageSchema:    r.UsageSchema(),
	}
}
