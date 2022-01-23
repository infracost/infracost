package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticache_replication_group",
		RFunc: NewElastiCacheReplicationGroup,
	}
}
func NewElastiCacheReplicationGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ElastiCacheReplicationGroup{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		NodeType:                    d.Get("node_type").String(),
		Engine:                      d.Get("engine").String(),
		CacheClusters:               d.Get("number_cache_clusters").Int(),
		ClusterDisabled:             d.Get("cluster_enabled").Type != gjson.Null && !d.Get("cluster_enabled").Bool(),
		ClusterMode:                 d.Get("cluster_mode").String(),
		ClusterNodeGroups:           d.Get("cluster_mode.0.num_node_groups").Int(),
		ClusterReplicasPerNodeGroup: d.Get("cluster_mode.0.replicas_per_node_group").Int(),
		SnapshotRetentionLimit:      d.Get("snapshot_retention_limit").Int(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
