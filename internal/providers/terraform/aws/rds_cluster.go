package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster",
		RFunc: NewRdsCluster,
	}
}
func NewRdsCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RdsCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), Engine: strPtr(d.Get("engine").String()), BackupRetentionPeriod: intPtr(d.Get("backup_retention_period").Int())}
	if !d.IsEmpty("engine_mode") {
		r.EngineMode = strPtr(d.Get("engine_mode").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
