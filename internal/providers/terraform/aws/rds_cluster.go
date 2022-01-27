package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster",
		RFunc: NewRDSCluster,
	}
}

func NewRDSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RDSCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		Engine:                d.Get("engine").String(),
		BackupRetentionPeriod: d.Get("backup_retention_period").Int(),
		EngineMode:            d.Get("engine_mode").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
