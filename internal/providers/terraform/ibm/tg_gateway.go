package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getTgGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_tg_gateway",
		RFunc: newTgGateway,
	}
}

func newTgGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	globalRouting := d.Get("global").Bool()
	r := &ibm.TgGateway{
		Address:       d.Address,
		Region:        region,
		GlobalRouting: globalRouting,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["globalRouting"] = globalRouting

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
