package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"

	"strings"
)

func getIsInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_instance",
		RFunc: newIsInstance,
	}
}

// valid profile values https://cloud.ibm.com/docs/vpc?topic=vpc-profiles&interface=ui
// profile names in Global Catalog contain dots instead of dashes
func newIsInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	profile := d.Get("profile").String()
	zone := d.Get("zone").String()
	truncatedZone := strings.Join(strings.Split(zone, "-")[0:2], "-") // the last part of the zone is dropped (eg: us-south-1 -> us-south)
	dedicatedHost := strings.TrimSpace(d.Get("dedicated_host").String())
	dedicatedHostGroup := strings.TrimSpace(d.Get("dedicated_host_group").String())
	isDedicated := !((dedicatedHost == "") && (dedicatedHostGroup == ""))

	SetCatalogMetadata(d, d.Type)

	r := &ibm.IsInstance{
		Address:       d.Address,
		Region:        region,
		Profile:       profile,
		Zone:          zone,
		TruncatedZone: truncatedZone,
		IsDedicated:   isDedicated,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
