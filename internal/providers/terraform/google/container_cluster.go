package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getContainerClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_container_cluster",
		RFunc: newContainerCluster,
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
		},
	}
}

func newContainerCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	isZone := isZone(region)

	if isZone {
		region = zoneToRegion(region)
	}

	autopilotEnabled := d.Get("enable_autopilot").Bool()

	var defaultNodePool *google.ContainerNodePool
	nodePools := make([]*google.ContainerNodePool, 0)

	if !d.Get("remove_default_node_pool").Bool() && !autopilotEnabled {
		zones := int64(zoneCount(d.RawValues, ""))

		countPerZone := int64(3)

		if !d.IsEmpty("initial_node_count") {
			countPerZone = d.Get("initial_node_count").Int()
		}

		defaultNodePool = &google.ContainerNodePool{
			Address:      "default_pool",
			Region:       region,
			Zones:        zones,
			CountPerZone: countPerZone,
			NodeConfig:   newContainerNodeConfig(d.Get("node_config.0")),
		}
	}

	if !autopilotEnabled {
		for i, values := range d.Get("node_pool").Array() {
			nodePool := newNodePool(fmt.Sprintf("node_pool[%d]", i), values, d)
			if nodePool != nil {
				nodePools = append(nodePools, nodePool)
			}
		}
	}

	r := &google.ContainerCluster{
		Address:          d.Address,
		Region:           region,
		AutopilotEnabled: autopilotEnabled,
		IsZone:           isZone,
		DefaultNodePool:  defaultNodePool,
		NodePools:        nodePools,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}

func zoneCount(d gjson.Result, location string) int {
	if location == "" {
		location = d.Get("location").String()
	}

	c := 3

	if isZone(location) {
		c = 1
	}

	if len(d.Get("node_locations").Array()) > 0 {
		c = len(d.Get("node_locations").Array())
	}

	return c
}

func newContainerNodeConfig(d gjson.Result) *google.ContainerNodeConfig {
	machineType := "e2-medium"
	if d.Get("machine_type").Exists() {
		machineType = d.Get("machine_type").String()
	}

	purchaseOption := "on_demand"
	if d.Get("preemptible").Bool() {
		purchaseOption = "preemptible"
	}

	diskType := "pd-standard"
	if d.Get("disk_type").Exists() {
		diskType = d.Get("disk_type").String()
	}

	diskSize := int64(100)
	if d.Get("disk_size_gb").Exists() {
		diskSize = d.Get("disk_size_gb").Int()
	}

	localSSDCount := d.Get("local_ssd_count").Int()

	guestAccelerators := collectComputeGuestAccelerators(d.Get("guest_accelerator"))

	return &google.ContainerNodeConfig{
		MachineType:       machineType,
		PurchaseOption:    purchaseOption,
		DiskType:          diskType,
		DiskSize:          float64(diskSize),
		LocalSSDCount:     localSSDCount,
		GuestAccelerators: guestAccelerators,
	}
}
