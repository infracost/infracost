package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster",
		RFunc: NewDocdbCluster,
	}

}
func NewDocdbCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocdbCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("backup_retention_period") {
		r.BackupRetentionPeriod = intPtr(d.Get("backup_retention_period").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
