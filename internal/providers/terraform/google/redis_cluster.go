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
		MemorySizeGB:	d.Get("memory_size_gb").Float(),
		Tier:			d.Get("tier").String(),
	}
}
