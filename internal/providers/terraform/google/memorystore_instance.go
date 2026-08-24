package google

import (
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getMemorystoreInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_memorystore_instance",
		CoreRFunc: NewMemorystoreInstance,
	}
}

func NewMemorystoreInstance(d *schema.ResourceData) schema.CoreResource {
	nodeTypeValue := d.Get("node_type")
	if nodeTypeValue.Exists() && nodeTypeValue.Type == gjson.Null {
		logging.Logger.Warn().Msgf("Skipping resource %s. Unknown node_type", d.Address)
		return nil
	}

	nodeType := strings.ToUpper(d.GetStringOrDefault("node_type", "HIGHMEM_MEDIUM"))
	nodeCapacityGB, ok := memorystoreNodeCapacityGB(nodeType)
	if !ok {
		logging.Logger.Warn().Msgf("Skipping resource %s. Unknown node_type %s", d.Address, nodeType)
		return nil
	}

	var nodeCount *int64
	var aofProvisionedGB *float64
	if !d.IsEmpty("shard_count") {
		shardCount := d.Get("shard_count").Int()
		aofProvisionedGBValue := float64(shardCount) * nodeCapacityGB
		aofProvisionedGB = &aofProvisionedGBValue

		if !d.Get("replica_count").Exists() || !d.IsEmpty("replica_count") {
			replicaCount := d.GetInt64OrDefault("replica_count", 0)
			nodeCountValue := shardCount * (1 + replicaCount)
			nodeCount = &nodeCountValue
		}
	}

	return &google.MemorystoreInstance{
		Address:          d.Address,
		Region:           d.Get("location").String(),
		NodeType:         nodeType,
		NodeCount:        nodeCount,
		AOFEnabled:       strings.EqualFold(d.Get("persistence_config.0.mode").String(), "AOF"),
		AOFProvisionedGB: aofProvisionedGB,
		BackupsEnabled:   len(d.Get("automated_backup_config").Array()) > 0,
	}
}

func memorystoreNodeCapacityGB(nodeType string) (float64, bool) {
	switch nodeType {
	case "SHARED_CORE_NANO":
		return 1.4, true
	case "CUSTOM_PICO":
		return 1.25, true
	case "CUSTOM_MICRO":
		return 2.5, true
	case "CUSTOM_MINI":
		return 3.75, true
	case "STANDARD_SMALL":
		return 6.5, true
	case "HIGHMEM_MEDIUM", "HIGHCPU_MEDIUM":
		return 13, true
	case "STANDARD_LARGE":
		return 26, true
	case "HIGHMEM_XLARGE":
		return 58, true
	case "HIGHMEM_2XLARGE":
		return 110, true
	default:
		return 0, false
	}
}
