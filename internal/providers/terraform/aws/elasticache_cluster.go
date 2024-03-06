package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_elasticache_cluster",
		CoreRFunc:           NewElastiCacheCluster,
		ReferenceAttributes: []string{"replication_group_id"},
	}
}

func NewElastiCacheCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.ElastiCacheCluster{
		Address:                d.Address,
		Region:                 d.Get("region").String(),
		NodeType:               d.Get("node_type").String(),
		Engine:                 d.Get("engine").String(),
		CacheNodes:             d.Get("num_cache_nodes").Int(),
		SnapshotRetentionLimit: d.Get("snapshot_retention_limit").Int(),
	}
	return r
}
