package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getIbmCosBucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_cos_bucket",
		RFunc:               newIbmCosBucket,
		ReferenceAttributes: []string{"resource_instance_id"},
	}
}

func getLocation(d *schema.ResourceData) (string, string) {
	if d.Get("region_location").Type != gjson.Null {
		return d.Get("region_location").String(), "region_location"
	}
	if d.Get("single_site_location").Type != gjson.Null {
		return d.Get("single_site_location").String(), "single_site_location"
	}
	if d.Get("cross_region_location").Type != gjson.Null {
		return d.Get("cross_region_location").String(), "cross_region_location"
	}
	return "", ""
}

func newIbmCosBucket(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	l, i := getLocation(d)

	r := &ibm.IbmCosBucket{
		Address:            d.Address,
		Region:             d.Get("region").String(),
		Location:           l,
		LocationIdentifier: i,
		StorageClass:       d.Get("storage_class").String(),
	}

	r.PopulateUsage(u)

	return r.BuildResource()
}
