package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_rds_cluster",
		CoreRFunc: NewRDSCluster,
	}
}

func NewRDSCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.RDSCluster{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		Engine:                d.GetStringOrDefault("engine", "aurora"),
		BackupRetentionPeriod: d.GetInt64OrDefault("backup_retention_period", 1),
		EngineMode:            d.GetStringOrDefault("engine_mode", "provisioned"),
	}
	return r
}
