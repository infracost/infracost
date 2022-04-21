package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_address",
		RFunc:               newComputeAddress,
		ReferenceAttributes: []string{},
	}
}
func getComputeGlobalAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_global_address",
		RFunc:               newComputeAddress,
		ReferenceAttributes: []string{},
	}
}

func newComputeAddress(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeAddress{
		Address:     d.Address,
		Region:      d.Get("region").String(),
		AddressType: d.Get("address_type").String(),
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
