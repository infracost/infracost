package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_rds_cluster",
		RFunc: NewRDSCluster,
	}
}

func NewRDSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	engineMode := d.GetStringOrDefault("engineMode", "provisioned")
	r := &aws.RDSCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		Engine:                d.GetStringOrDefault("engine", "aurora"),
		BackupRetentionPeriod: d.GetInt64OrDefault("backupRetentionPeriod", 1),
		EngineMode:            engineMode,
		IOOptimized:           d.Get("storageType").String() == "aurora-iopt1",
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
