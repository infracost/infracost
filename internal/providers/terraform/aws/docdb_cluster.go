package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_docdb_cluster",
		CoreRFunc: NewDocDBCluster,
	}

}
func NewDocDBCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.DocDBCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		BackupRetentionPeriod: d.Get("backup_retention_period").Int(),
	}
	return r
}
