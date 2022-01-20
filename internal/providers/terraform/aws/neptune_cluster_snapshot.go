package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getNeptuneClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster_snapshot",
		RFunc: NewNeptuneClusterSnapshot,
		ReferenceAttributes: []string{
			"db_cluster_identifier",
		},
	}
}

func NewNeptuneClusterSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var backupRetentionPeriod *int64

	dbClusterIdentifiers := d.References("db_cluster_identifier")
	if len(dbClusterIdentifiers) > 0 {
		cluster := dbClusterIdentifiers[0]
		if cluster.Get("backup_retention_period").Type != gjson.Null {
			backupRetentionPeriod = intPtr(cluster.Get("backup_retention_period").Int())
		}
	}

	r := &aws.NeptuneClusterSnapshot{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		BackupRetentionPeriod: backupRetentionPeriod,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
