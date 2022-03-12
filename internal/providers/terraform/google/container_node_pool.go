package google

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetContainerNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_container_node_pool",
		RFunc: NewContainerNodePool,
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

func NewContainerNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var cluster *schema.ResourceData
	if len(d.References("cluster")) > 0 {
		cluster = d.References("cluster")[0]
	}

	var countPerZoneOverride *int64
	if u != nil && u.Get("nodes").Exists() {
		c := u.Get("nodes").Int()
		countPerZoneOverride = &c
	}

	return newNodePool(d.Address, d.RawValues, countPerZoneOverride, cluster)
}

func newNodePool(address string, d gjson.Result, countPerZoneOverride *int64, cluster *schema.ResourceData) *schema.Resource {
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

	if d.Get("initial_node_count").Type != gjson.Null {
		countPerZone = d.Get("initial_node_count").Int()
	}

	if d.Get("node_count").Type != gjson.Null {
		countPerZone = d.Get("node_count").Int()
	}

	if d.Get("autoscaling.0.min_node_count").Type != gjson.Null {
		countPerZone = d.Get("autoscaling.0.min_node_count").Int()
	}

	if countPerZoneOverride != nil {
		countPerZone = *countPerZoneOverride
	}

	nodeCount := decimal.NewFromInt(zones * countPerZone)

	r := &schema.Resource{
		Name:           address,
		CostComponents: nodePoolCostComponents(region, d.Get("node_config.0")),
	}

	schema.MultiplyQuantities(r, nodeCount)

	return r
}

func nodePoolCostComponents(region string, nodeConfig gjson.Result) []*schema.CostComponent {
	poolSize := decimal.NewFromInt(1)
	machineType := "e2-medium"
	if nodeConfig.Get("machine_type").Exists() {
		machineType = nodeConfig.Get("machine_type").String()
	}

	purchaseOption := "on_demand"
	if nodeConfig.Get("preemptible").Bool() {
		purchaseOption = "preemptible"
	}

	diskType := "pd-standard"
	if nodeConfig.Get("disk_type").Exists() {
		diskType = nodeConfig.Get("disk_type").String()
	}

	diskSize := decimal.NewFromInt(100)
	if nodeConfig.Get("disk_size_gb").Exists() {
		diskSize = decimal.NewFromInt(nodeConfig.Get("disk_size_gb").Int())
	}

	costComponents := []*schema.CostComponent{
		computeCostComponent(region, machineType, purchaseOption, poolSize),
		computeDisk(region, diskType, &diskSize, poolSize),
	}

	localSSDCount := nodeConfig.Get("local_ssd_count").Int()
	if localSSDCount > 0 {
		costComponents = append(costComponents, scratchDisk(region, purchaseOption, int(localSSDCount)))
	}

	for _, guestAccel := range nodeConfig.Get("guest_accelerator").Array() {
		costComponents = append(costComponents, guestAccelerator(region, purchaseOption, guestAccel, poolSize))
	}

	return costComponents
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
