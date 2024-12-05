package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeRegionInstanceGroupManagerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_region_instance_group_manager",
		CoreRFunc:           newComputeRegionInstanceGroupManager,
		Notes:               []string{"Multiple versions are not supported."},
		ReferenceAttributes: []string{"version.0.instance_template", "google_compute_region_per_instance_config.region_instance_group_manager"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			return []string{d.Get("name").String()}
		},
	}
}

func newComputeRegionInstanceGroupManager(d *schema.ResourceData) schema.CoreResource {
	targetSize := int64(1)
	if d.Get("target_size").Exists() {
		targetSize = d.Get("target_size").Int()
	}
	if len(d.References("google_compute_region_per_instance_config.region_instance_group_manager")) > 0 {
		targetSize += int64(len(d.References("google_compute_region_per_instance_config.region_instance_group_manager")))
	}

	var machineType string
	purchaseOption := "on_demand"
	scratchDisks := 0
	disks := []*google.ComputeDisk{}
	guestAccelerators := []*google.ComputeGuestAccelerator{}

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate := d.References("version.0.instance_template")[0]

		machineType = instanceTemplate.Get("machine_type").String()

		if instanceTemplate.Get("scheduling.0.preemptible").Bool() {
			purchaseOption = "preemptible"
		}

		for i, disk := range instanceTemplate.Get("disk").Array() {
			diskType := disk.Get("type").String()
			switch diskType {
			case "SCRATCH":
				scratchDisks++
			default:
				diskSize := int64(100)
				if size := disk.Get("disk_size_gb"); size.Exists() {
					diskSize = size.Int()
				}
				diskType := disk.Get("disk_type").String()

				disks = append(disks, &google.ComputeDisk{
					Address:       fmt.Sprintf("disk[%d]", i),
					Type:          diskType,
					Size:          float64(diskSize),
					Region:        d.Region,
					InstanceCount: &targetSize,
				})
			}
		}

		guestAccelerators = collectComputeGuestAccelerators(instanceTemplate.Get("guest_accelerator"))
	}

	r := &google.ComputeRegionInstanceGroupManager{
		Address:           d.Address,
		Region:            d.Region,
		MachineType:       machineType,
		PurchaseOption:    purchaseOption,
		TargetSize:        targetSize,
		Disks:             disks,
		ScratchDisks:      scratchDisks,
		GuestAccelerators: guestAccelerators,
	}
	return r
}
