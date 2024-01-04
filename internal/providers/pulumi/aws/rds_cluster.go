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
		Engine:                d.GetStringOrDefault("engine", "aurora"),
		BackupRetentionPeriod: d.GetInt64OrDefault("backupRetentionPeriod", 1),
		EngineMode:            d.GetStringOrDefault("engineMode", "provisioned"),
		StorageGB:             floatPtr(d.GetFloat64OrDefault("allocatedStorage", 0)),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
