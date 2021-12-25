package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetDocDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster",
		RFunc: NewDocDBCluster,
	}

}
func NewDocDBCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocDBCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("backup_retention_period").Exists() && d.Get("backup_retention_period").Type != gjson.Null {
		r.BackupRetentionPeriod = intPtr(d.Get("backup_retention_period").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
