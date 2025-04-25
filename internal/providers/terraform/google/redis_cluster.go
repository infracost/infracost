package google

import (
	"fmt"
	"strings"
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
	aofMode := strings.ToLower(d.Get("persistence_config.0.mode").String())
	automatedBackups := d.Get("automated_backup_config").Array()

	provisionedGB := d.Get("memory_size_gb").Int()

	if provisionedGB == 0 {
		estimatedGB := estimateMemorySizeByNodeType(d.Get("node_type").String())
		if estimatedGB > 0 {
			provisionedGB = estimatedGB
			fmt.Printf("[Warning] memory_size_gb for Redis cluster '%s' was not provided, estimated as %d GB based on node_type '%s'.\n", d.Address, provisionedGB, d.Get("node_type").String())
		} else {
			fmt.Printf("[Warning] memory_size_gb for Redis cluster '%s' could not be estimated (unknown node_type '%s').\n", d.Address, d.Get("node_type").String())
		}
	}

	return &google.RedisCluster{
		Address:		d.Address,
		Region:			d.Get("region").String(),
		NodeType:	    d.Get("node_type").String(),
		NodeCount:	    int(d.Get("shard_count").Int() * (1 + d.Get("replica_count").Int())),
		ProvisionedGB:	provisionedGB,
		AOFEnabled:		aofMode == "aof",
		BackupsEnabled:	len(automatedBackups) > 0,
	}
}

func estimateMemorySizeByNodeType(nodeType string) int64 {
	switch strings.ToUpper(nodeType) {
	case "REDIS_SHARED_CORE_NANO":
		return 2 // Slightly rounded from 1.4 GB to 2 GB (safer estimation)
	case "REDIS_STANDARD_SMALL":
		return 7 // 6.5 GB rounded up
	case "REDIS_HIGHMEM_MEDIUM":
		return 13
	case "REDIS_HIGHMEM_XLARGE":
		return 58
	default:
		return 0 // Unknown node_type
	}
}