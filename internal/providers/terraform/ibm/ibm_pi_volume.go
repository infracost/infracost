package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIbmPiVolumeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_pi_volume",
		RFunc: newIbmPiVolume,
	}
}

func newIbmPiVolume(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &ibm.IbmPiVolume{
		Address:          d.Address,
		Region:           region,
		Name:             d.Get("pi_volume_name").String(),
		Size:             d.Get("pi_volume_size").Int(),
		Type:             d.Get("pi_volume_type").String(),
		VolumePool:       d.Get("pi_volume_pool").String(),
		AffinityPolicy:   d.Get("pi_affinity_policy").String(),
		AffinityInstance: d.Get("pi_affinity_instance").String(),
		AffinityVolume:   d.Get("pi_affinity_volume").String(),
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["profile"] = r.Type
	configuration["capacity"] = r.Size

	if r.VolumePool != "" {
		configuration["volume_pool"] = r.VolumePool
	}
	if r.AffinityPolicy != "" {
		configuration["affinity"] = r.AffinityPolicy
	}
	if r.AffinityInstance != "" {
		configuration["affinity_instance"] = r.AffinityInstance
	}
	if r.AffinityVolume != "" {
		configuration["affinity_volume"] = r.AffinityVolume
	}

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
