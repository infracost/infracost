package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getRedisClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:	"google_redis_cluster",
		CoreRFunc: NewRedisCluster,
	}
}

func NewRedisCluster(d *schema.ResourceData) schema.CoreResource {
	return &google.RedisCluster{
		Address:		d.Address,
		Region:			d.Get("region").String(),
		NodeType:	    d.Get("node_type").String(),
		NodeCount:	    int(d.Get("shard_count").Int() * (1 + d.Get("replica_count").Int())),
	}
}
