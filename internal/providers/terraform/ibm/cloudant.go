package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudantRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_cloudant",
		RFunc: newCloudant,
	}
}

func newCloudant(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	p := d.Get("plan").String()
	c := d.Get("capacity").String()

	r := &ibm.Cloudant{
		Address:  d.Address,
		Region:   region,
		Plan:     p,
		Capacity: c,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["plan"] = p
	configuration["capacity"] = c

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
