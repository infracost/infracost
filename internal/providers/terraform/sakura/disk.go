package sakura

import (
	"github.com/infracost/infracost/internal/resources/sakura"
	"github.com/infracost/infracost/internal/schema"
)

func getDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "sakura_disk",
		CoreRFunc: NewDisk,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return GetResourceRegion(d)
		},
	}
}

func NewDisk(d *schema.ResourceData) schema.CoreResource {
	return &sakura.Disk{
		Address: d.Address,
		Zone:    d.Get("zone").String(),
		Plan:    d.Get("plan").String(),
		Size:    d.Get("size").Int(),
	}
}
