package google

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_container_node_pool",
		RFunc: newContainerNodePool,
		ReferenceAttributes: []string{
			"cluster",
		},
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			var location string

			var cluster *schema.ResourceData
			if len(d.References("cluster")) > 0 {
				cluster = d.References("cluster")[0]
			}

			if cluster != nil {
				location = cluster.Get("location").String()
			}

			if d.Get("location").String() != "" {
				location = d.Get("location").String()
			}

			region := location
			if isZone(location) {
				region = zoneToRegion(location)
			}

			return region
		},
	}
}

func newContainerNodePool(d *schema.ResourceData) schema.CoreResource {
	var cluster *schema.ResourceData
	if len(d.References("cluster")) > 0 {
		cluster = d.References("cluster")[0]
	}

	r := newNodePool(d.Address, d.RawValues, cluster)

	if r == nil {
		return nil
	}

	return r
}

func newNodePool(address string, d gjson.Result, cluster *schema.ResourceData) *google.ContainerNodePool {
	var location string

	if cluster != nil {
		location = cluster.Get("location").String()
	}

	if d.Get("location").String() != "" {
		location = d.Get("location").String()
	}

	region := location
	if isZone(location) {
		region = zoneToRegion(location)
	}

	if region == "" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Unable to determine region", address)
		return nil
	}

	zones := int64(3)

	if cluster != nil {
		zones = int64(zoneCount(cluster.RawValues, ""))
	}

	if len(d.Get("nodeLocations").Array()) > 0 {
		zones = int64(zoneCount(d, location))
	}

	countPerZone := int64(3)

	if d.Get("initialNodeCount").Exists() {
		countPerZone = d.Get("initialNodeCount").Int()
	}

	if d.Get("nodeCount").Exists() {
		countPerZone = d.Get("nodeCount").Int()
	}

	if d.Get("autoscaling.0.minNodeCount").Exists() {
		countPerZone = d.Get("autoscaling.0.minNodeCount").Int()
	}

	containerNodeConfig := newContainerNodeConfig(d.Get("nodeConfig.0"))

	return &google.ContainerNodePool{
		Address:      address,
		Region:       region,
		Zones:        zones,
		CountPerZone: countPerZone,
		NodeConfig:   containerNodeConfig,
	}
}
