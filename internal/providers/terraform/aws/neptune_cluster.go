package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNeptuneClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_neptune_cluster",
		CoreRFunc: NewNeptuneCluster,
	}
}

func NewNeptuneCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.NeptuneCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		BackupRetentionPeriod: d.Get("backup_retention_period").Int(),
	}
	return r
}
