package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNeptuneClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_neptune_cluster_snapshot",
		CoreRFunc: NewNeptuneClusterSnapshot,
		ReferenceAttributes: []string{
			"db_cluster_identifier",
		},
	}
}

func NewNeptuneClusterSnapshot(d *schema.ResourceData) schema.CoreResource {
	var backupRetentionPeriod *int64

	dbClusterIdentifiers := d.References("db_cluster_identifier")
	if len(dbClusterIdentifiers) > 0 {
		cluster := dbClusterIdentifiers[0]
		backupRetentionPeriod = intPtr(cluster.GetInt64OrDefault("backup_retention_period", 1))
	}

	r := &aws.NeptuneClusterSnapshot{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		BackupRetentionPeriod: backupRetentionPeriod,
	}
	return r
}
