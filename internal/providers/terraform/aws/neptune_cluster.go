package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetNeptuneClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster",
		RFunc: NewNeptuneCluster,
	}
}
func NewNeptuneCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.NeptuneCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("backup_retention_period").Exists() && d.Get("backup_retention_period").Type != gjson.Null {
		r.BackupRetentionPeriod = intPtr(d.Get("backup_retention_period").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
