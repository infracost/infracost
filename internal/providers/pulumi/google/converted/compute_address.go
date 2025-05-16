package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_address",
		RFunc: newComputeAddress,
		ReferenceAttributes: []string{
			"google_compute_instance.network_interface.0.access_config.0.nat_ip",
		},
	}
}
func getComputeGlobalAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_global_address",
		RFunc: newComputeAddress,
		ReferenceAttributes: []string{
			"google_compute_instance.network_interface.0.access_config.0.nat_ip",
		},
	}
}

func newComputeAddress(d *schema.ResourceData) schema.CoreResource {
	purchaseOption := ""
	instanceRefs := d.References("googleComputeInstance.networkInterface.0.accessConfig.0.natIp")
	if len(instanceRefs) > 0 {
		purchaseOption = getComputePurchaseOption(instanceRefs[0].RawValues)
	}

	r := &google.ComputeAddress{
		Address:                d.Address,
		Region:                 d.Get("region").String(),
		AddressType:            d.Get("addressType").String(),
		Purpose:                d.Get("purpose").String(),
		InstancePurchaseOption: purchaseOption,
	}
	return r
}
