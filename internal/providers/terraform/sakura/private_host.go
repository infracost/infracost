package sakura

import (
	"github.com/infracost/infracost/internal/resources/sakura"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateHostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "sakura_private_host",
		CoreRFunc: NewPrivateHost,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return GetResourceRegion(d)
		},
	}
}

func NewPrivateHost(d *schema.ResourceData) schema.CoreResource {
	return &sakura.PrivateHost{
		Address:            d.Address,
		Zone:               d.Get("zone").String(),
		Class:              d.Get("class").String(),
		DedicatedStorageID: d.Get("dedicated_storage_id").String(),
	}
}
