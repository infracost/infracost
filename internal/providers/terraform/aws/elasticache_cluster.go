package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_elasticache_cluster",
		CoreRFunc: NewElastiCacheCluster,
		ReferenceAttributes: []string{
			"replication_group_id",
		},
	}
}

func NewElastiCacheCluster(d *schema.ResourceData) schema.CoreResource {
	// If using Terraform plan, the replication_group_id attribute can be empty even if it has a reference
	// so check if the references as well.
	replicationGroupRefs := d.References("replication_group_id")
	hasReplicationGroup := len(replicationGroupRefs) > 0 || !d.IsEmpty("replication_group_id")

	r := &aws.ElastiCacheCluster{
		Address:                d.Address,
		Region:                 d.Get("region").String(),
		NodeType:               d.Get("node_type").String(),
		Engine:                 d.Get("engine").String(),
		CacheNodes:             d.Get("num_cache_nodes").Int(),
		SnapshotRetentionLimit: d.Get("snapshot_retention_limit").Int(),
		HasReplicationGroup:    hasReplicationGroup,
	}
	return r
}
