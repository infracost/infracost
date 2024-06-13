package google

import (
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_container_cluster",
		CoreRFunc: newContainerCluster,
		// this is a reverse reference, it depends on the container_node_pool RegistryItem
		// defining "cluster" as a ReferenceAttribute
		ReferenceAttributes: []string{"google_container_node_pool.cluster"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			name := d.Get("name").String()
			if name != "" {
				return []string{name}
			}

			return nil
		},
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := d.Get("location").String()
			isZone := isZone(region)

			if isZone {
				region = zoneToRegion(region)
			}

			return region
		},
	}
}

func newContainerCluster(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	isZone := isZone(region)

	autopilotEnabled := d.Get("enable_autopilot").Bool()

	var defaultNodePool *google.ContainerNodePool
	nodePools := make([]*google.ContainerNodePool, 0)

	// Build a list of node pools that are defined in other `google_container_node_pool` resources
	// If we find these in the `node_pool` field, we want to skip them for this resource
	// since we will show the cost for these against their `google_container_node_pool` resource
	definedNodePoolNames := []string{}
	for _, ref := range d.References("google_container_node_pool.cluster") {
		definedNodePoolNames = append(definedNodePoolNames, ref.Get("name").String())
	}

	if !autopilotEnabled {
		nameIndex := 0
		for _, values := range d.Get("node_pool").Array() {
			if contains(definedNodePoolNames, values.Get("name").String()) {
				logging.Logger.Debug().Msgf("Skipping node pool with name %s since it is defined in another resource", values.Get("name").String())
				continue
			}

			if values.Get("name").String() == "default-pool" {
				defaultNodePool = newNodePool("default_pool", values, d)
				continue
			}

			name := fmt.Sprintf("node_pool[%d]", nameIndex)
			nameIndex++

			nodePool := newNodePool(name, values, d)
			if nodePool != nil {
				nodePools = append(nodePools, nodePool)
			}
		}

		// Create the default pool if it isn't specified in the existing `node_pools` - this happens if the Terraform resources have not been applied.
		if defaultNodePool == nil && !d.Get("remove_default_node_pool").Bool() {
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
	}

	r := &google.ContainerCluster{
		Address:          d.Address,
		Region:           region,
		AutopilotEnabled: autopilotEnabled,
		IsZone:           isZone,
		DefaultNodePool:  defaultNodePool,
		NodePools:        nodePools,
	}
	return r
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
	if d.Get("preemptible").Bool() || d.Get("spot").Bool() {
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
