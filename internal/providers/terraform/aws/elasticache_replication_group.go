package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticache_replication_group",
		RFunc: NewElastiCacheReplicationGroup,
	}
}

func NewElastiCacheReplicationGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	nodeType := d.Get("node_type").String()
	var cacheNodes decimal.Decimal
	cacheEngine := "redis"

	if d.Get("engine").Exists() {
		cacheEngine = d.Get("engine").String()
	}

	if d.Get("cluster_mode").Exists() {
		nodeGroups := decimal.NewFromInt(d.Get("cluster_mode.0.num_node_groups").Int())
		shards := decimal.NewFromInt(d.Get("cluster_mode.0.replicas_per_node_group").Int())
		cacheNodes = nodeGroups.Mul(shards).Add(nodeGroups)
	} else {
		cacheNodes = decimal.NewFromInt(d.Get("number_cache_clusters").Int())
	}

	return newElasticacheResource(d, u, nodeType, cacheNodes, cacheEngine)
}
