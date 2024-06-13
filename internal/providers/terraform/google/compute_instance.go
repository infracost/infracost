package google

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_instance",
		CoreRFunc: newComputeInstance,
		ReferenceAttributes: []string{
			"network_interface.0.access_config.0.nat_ip", // google_compute_address
		},
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
			"Sole-tenant VMs are not supported.",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := d.Get("region").String()

			zone := d.Get("zone").String()
			if zone != "" {
				region = zoneToRegion(zone)
			}

			return region
		},
	}
}

func newComputeInstance(d *schema.ResourceData) schema.CoreResource {
	machineType := d.Get("machine_type").String()

	region := d.Region
	size := int64(1)

	purchaseOption := getComputePurchaseOption(d.RawValues)

	initializeParams := d.Get("boot_disk.0.initialize_params.0")
	bootDiskSize := float64(defaultVolumeSize)
	var bootDiskType string
	hasBootDisk := initializeParams.Exists()
	if hasBootDisk {
		if initializeParams.Get("size").Exists() {
			bootDiskSize = initializeParams.Get("size").Float()
		}

		bootDiskType = initializeParams.Get("type").String()
	}

	scratchDisks := len(d.Get("scratch_disk").Array())
	guestAccelerators := collectComputeGuestAccelerators(d.Get("guest_accelerator"))

	r := &google.ComputeInstance{
		Address:           d.Address,
		Region:            region,
		MachineType:       machineType,
		PurchaseOption:    purchaseOption,
		Size:              size,
		HasBootDisk:       hasBootDisk,
		BootDiskSize:      bootDiskSize,
		BootDiskType:      bootDiskType,
		ScratchDisks:      scratchDisks,
		GuestAccelerators: guestAccelerators,
	}
	return r
}

// getComputePurchaseOption determines the purchase option for Compute
// resources.
func getComputePurchaseOption(d gjson.Result) string {
	purchaseOption := "on_demand"
	if d.Get("scheduling.0.preemptible").Bool() {
		purchaseOption = "preemptible"
	}

	return purchaseOption
}

// collectComputeGuestAccelerators collects Guest Accelerator data for Compute
// resources.
func collectComputeGuestAccelerators(d gjson.Result) []*google.ComputeGuestAccelerator {
	guestAccelerators := []*google.ComputeGuestAccelerator{}

	for _, guestAccel := range d.Array() {
		guestAccelerators = append(guestAccelerators, &google.ComputeGuestAccelerator{
			Type:  guestAccel.Get("type").String(),
			Count: guestAccel.Get("count").Int(),
		})
	}

	return guestAccelerators
}
