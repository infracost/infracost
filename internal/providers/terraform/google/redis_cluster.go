package google

import (
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getRedisClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_redis_cluster",
		CoreRFunc: NewRedisCluster,
	}
}

func NewRedisCluster(d *schema.ResourceData) schema.CoreResource {
	aofMode := strings.ToLower(d.Get("persistence_config.0.mode").String())
	automatedBackups := d.Get("automated_backup_config").Array()

	provisionedGB := getAOFProvisionedGB(d.Get("node_type").String())
	if provisionedGB == 0 {
		logging.Logger.Warn().Msgf("Skipping resource %s. Unknown node_type %s", d.Address, d.Get("node_type").String())
		return nil
	}

	return &google.RedisCluster{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		NodeType:         d.Get("node_type").String(),
		NodeCount:        int(d.Get("shard_count").Int() * (1 + d.Get("replica_count").Int())),
		AOFProvisionedGB: provisionedGB,
		AOFEnabled:       aofMode == "aof",
		BackupsEnabled:   len(automatedBackups) > 0,
	}
}

func getAOFProvisionedGB(nodeType string) int64 {
	switch strings.ToUpper(nodeType) {
	case "REDIS_SHARED_CORE_NANO":
		return 2 // Rounded up from 1.4 GB
	case "REDIS_STANDARD_SMALL":
		return 7 // Rounded up from 6.5 GB
	case "REDIS_HIGHMEM_MEDIUM":
		return 13
	case "REDIS_HIGHMEM_XLARGE":
		return 58
	default:
		return 0 // Unknown node_type
	}
}
