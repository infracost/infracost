package sakura

import (
	"github.com/infracost/infracost/internal/resources/sakura"
	"github.com/infracost/infracost/internal/schema"
)

func getInternetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "sakura_internet",
		CoreRFunc: NewInternet,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return GetResourceRegion(d)
		},
	}
}

func NewInternet(d *schema.ResourceData) schema.CoreResource {
	return &sakura.Internet{
		Address:   d.Address,
		Zone:      d.Get("zone").String(),
		BandWidth: d.Get("band_width").Int(),
	}
}
