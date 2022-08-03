package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

const RegionLocation string = "region_location"
const SingleSiteLocation string = "single_site_location"
const CrossRegionLocation string = "cross_region_location"

func getIbmCosBucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_cos_bucket",
		RFunc:               newIbmCosBucket,
		ReferenceAttributes: []string{"resource_instance_id"},
	}
}

func getLocation(d *schema.ResourceData) (string, string) {
	if d.Get(RegionLocation).Type != gjson.Null {
		return d.Get(RegionLocation).String(), RegionLocation
	}
	if d.Get(SingleSiteLocation).Type != gjson.Null {
		return d.Get(SingleSiteLocation).String(), SingleSiteLocation
	}
	if d.Get(CrossRegionLocation).Type != gjson.Null {
		return d.Get(CrossRegionLocation).String(), CrossRegionLocation
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
