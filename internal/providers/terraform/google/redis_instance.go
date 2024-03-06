package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getRedisInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_redis_instance",
		CoreRFunc: NewRedisInstance,
	}
}

func NewRedisInstance(d *schema.ResourceData) schema.CoreResource {
	r := &google.RedisInstance{
		Address:      d.Address,
		Region:       d.Get("region").String(),
		MemorySizeGB: d.Get("memory_size_gb").Float(),
		Tier:         d.Get("tier").String(),
	}

	return r
}
