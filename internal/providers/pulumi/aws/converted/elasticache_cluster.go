package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_elasticache_cluster",
		RFunc: NewElastiCacheCluster,
		ReferenceAttributes: []string{
			"replication_group_id",
		},
	}
}

func NewElastiCacheCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	// If using Terraform plan, the replication_group_id attribute can be empty even if it has a reference
	// so check if the references as well.
	replicationGroupRefs := d.References("replicationGroupId")
	hasReplicationGroup := len(replicationGroupRefs) > 0 || !d.IsEmpty("replication_group_id")

	r := &aws.ElastiCacheCluster{
		Address:                d.Address,
		Region:                 d.Get("region").String(),
		NodeType:               d.Get("nodeType").String(),
		Engine:                 d.Get("engine").String(),
		CacheNodes:             d.Get("numCacheNodes").Int(),
		SnapshotRetentionLimit: d.Get("snapshotRetentionLimit").Int(),
		HasReplicationGroup:    hasReplicationGroup,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
