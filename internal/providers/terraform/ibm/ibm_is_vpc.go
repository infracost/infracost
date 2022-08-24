package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIbmIsVpcRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_vpc",
		RFunc: newIbmIsVpc,
	}
}

func newIbmIsVpc(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	r := &ibm.IbmIsVpc{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
