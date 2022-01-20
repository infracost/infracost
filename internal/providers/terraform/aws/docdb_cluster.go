package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster",
		RFunc: NewDocDBCluster,
	}

}
func NewDocDBCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocDBCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		BackupRetentionPeriod: d.Get("backup_retention_period").Int(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
