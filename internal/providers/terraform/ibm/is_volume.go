package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIsVolumeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_volume",
		RFunc: newIsVolume,
	}
}

func newIsVolume(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	capacity := d.Get("capacity").Int()
	if capacity == 0 {
		capacity = 100
	}
	r := &ibm.IsVolume{
		Address:  d.Address,
		Region:   region,
		Profile:  d.Get("profile").String(),
		IOPS:     d.Get("iops").Int(),
		Capacity: capacity,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
