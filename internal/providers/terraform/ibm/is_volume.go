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
	profile := d.Get("profile").String()
	iops := d.Get("iops").Int()

	if capacity == 0 {
		capacity = 100
	}
	r := &ibm.IsVolume{
		Address:  d.Address,
		Region:   region,
		Profile:  profile,
		IOPS:     iops,
		Capacity: capacity,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["profile"] = profile
	configuration["capacity"] = capacity
	configuration["iops"] = iops

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
