package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster",
		RFunc: NewRDSCluster,
	}
}
func NewRDSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RDSCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), Engine: strPtr(d.Get("engine").String()), BackupRetentionPeriod: intPtr(d.Get("backup_retention_period").Int())}
	if d.Get("engine_mode").Exists() && d.Get("engine_mode").Type != gjson.Null {
		r.EngineMode = strPtr(d.Get("engine_mode").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
