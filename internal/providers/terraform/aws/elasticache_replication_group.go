package aws

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticache_replication_group",
		RFunc: NewElastiCacheReplicationGroup,
	}
}

func NewElastiCacheReplicationGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	nodeType := d.Get("node_type").String()
	var cacheNodes decimal.Decimal
	cacheEngine := "redis"

	if d.Get("engine").Exists() {
		cacheEngine = d.Get("engine").String()
	}

	// the aws_elasticache_replication_group state output has cluster_mode set as
	//
	//   cluster_mode {
	//    replicas_per_node_group = 1
	//    num_node_groups         = 1
	//  }
	//
	// even when cluster mode is disabled. This causes an issue with infracost diff as it sees no change in prices
	// as both the new and the old resource have cluster_mode with 1 1 set. In order to circumvent this problem we
	// need to explicitly check that cluster_enabled attribute (output attribute) is set to false in the terraform state.
	// This will only be present in a state/diff run and won't be available in breakdown or output runs.
	clusterDisabled := d.Get("cluster_enabled").Type != gjson.Null && d.Get("cluster_enabled").Bool()
	if d.Get("cluster_mode").Exists() && !clusterDisabled {
		nodeGroups := decimal.NewFromInt(d.Get("cluster_mode.0.num_node_groups").Int())
		shards := decimal.NewFromInt(d.Get("cluster_mode.0.replicas_per_node_group").Int())
		cacheNodes = nodeGroups.Mul(shards).Add(nodeGroups)
	} else {
		cacheNodes = decimal.NewFromInt(d.Get("number_cache_clusters").Int())
	}

	return newElasticacheResource(d, u, nodeType, cacheNodes, cacheEngine)
}
