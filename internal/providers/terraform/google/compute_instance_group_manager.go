package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeInstanceGroupManagerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_instance_group_manager",
		RFunc:               NewComputeInstanceGroupManager,
		Notes:               []string{"Multiple versions are not supported."},
		ReferenceAttributes: []string{"version.0.instance_template"},
	}
}

func NewComputeInstanceGroupManager(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var region string
	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}

	machineType := getMachineType(d)
	purchaseOption := getPurchaseOption(d)

	targetSize := decimal.NewFromInt(1)
	if d.Get("target_size").Exists() {
		targetSize = decimal.NewFromInt(d.Get("target_size").Int())
	}

	costComponents := []*schema.CostComponent{computeCostComponent(region, machineType, purchaseOption, targetSize)}

	diskCostComponents := getDisksFromTemplate(d, region, targetSize)
	costComponents = append(costComponents, diskCostComponents...)

	guestAcceleratorComponents := getAcceleratorsFromTemplate(d, region, purchaseOption, targetSize)
	costComponents = append(costComponents, guestAcceleratorComponents...)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func getMachineType(d *schema.ResourceData) string {
	var instanceTemplate *schema.ResourceData
	var machineType string

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate = d.References("version.0.instance_template")[0]
	} else {
		return ""
	}

	if instanceTemplate.Get("machine_type").Exists() {
		machineType = instanceTemplate.Get("machine_type").String()
	}
	return machineType
}

func getPurchaseOption(d *schema.ResourceData) string {
	var instanceTemplate *schema.ResourceData
	purchaseOption := "on_demand"

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate = d.References("version.0.instance_template")[0]
	} else {
		return purchaseOption
	}

	if instanceTemplate.Get("scheduling.0.preemptible").Bool() {
		purchaseOption = "preemptible"
	}
	return purchaseOption
}

func getDisksFromTemplate(d *schema.ResourceData, region string, targetSize decimal.Decimal) []*schema.CostComponent {
	var instanceTemplate *schema.ResourceData
	var costComponents []*schema.CostComponent

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate = d.References("version.0.instance_template")[0]
	} else {
		return costComponents
	}

	if len(instanceTemplate.Get("disk").Array()) > 0 {
		for _, disk := range instanceTemplate.Get("disk").Array() {
			diskSize := decimal.NewFromInt(100)
			if size := disk.Get("disk_size_gb"); size.Exists() {
				diskSize = decimal.NewFromInt(size.Int())
			}
			diskType := disk.Get("disk_type").String()
			costComponents = append(costComponents, computeDisk(region, diskType, &diskSize, targetSize))
		}
	}

	return costComponents
}

func getAcceleratorsFromTemplate(d *schema.ResourceData, region, purchaseOption string, targetSize decimal.Decimal) []*schema.CostComponent {
	var instanceTemplate *schema.ResourceData
	var costComponents []*schema.CostComponent

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate = d.References("version.0.instance_template")[0]
	} else {
		return costComponents
	}

	for _, guestAccel := range instanceTemplate.Get("guest_accelerator").Array() {
		costComponents = append(costComponents, guestAccelerator(region, purchaseOption, guestAccel, targetSize))
	}

	return costComponents
}
