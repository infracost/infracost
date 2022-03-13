package google

import (
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_container_node_pool",
		RFunc: newContainerNodePool,
		ReferenceAttributes: []string{
			"cluster",
		},
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
		},
	}
}

func newContainerNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var cluster *schema.ResourceData
	if len(d.References("cluster")) > 0 {
		cluster = d.References("cluster")[0]
	}

	r := newNodePool(d.Address, d.RawValues, cluster)

	if r == nil {
		return nil
	}

	r.PopulateUsage(u)

	return r.BuildResource()
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
		log.Warnf("Skipping resource %s. Unable to determine region", address)
		return nil
	}

	zones := int64(3)

	if cluster != nil {
		zones = int64(zoneCount(cluster.RawValues, ""))
	}

	if len(d.Get("node_locations").Array()) > 0 {
		zones = int64(zoneCount(d, location))
	}

	countPerZone := int64(3)

	if d.Get("initial_node_count").Exists() {
		countPerZone = d.Get("initial_node_count").Int()
	}

	if d.Get("node_count").Exists() {
		countPerZone = d.Get("node_count").Int()
	}

	if d.Get("autoscaling.0.min_node_count").Exists() {
		countPerZone = d.Get("autoscaling.0.min_node_count").Int()
	}

	containerNodeConfig := newContainerNodeConfig(d.Get("node_config.0"))

	return &google.ContainerNodePool{
		Address:      address,
		Region:       region,
		Zones:        zones,
		CountPerZone: countPerZone,
		NodeConfig:   containerNodeConfig,
	}
}
