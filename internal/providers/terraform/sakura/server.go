package sakura

import (
	"github.com/infracost/infracost/internal/resources/sakura"
	"github.com/infracost/infracost/internal/schema"
)

func getServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "sakura_server",
		CoreRFunc: NewServer,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return GetResourceRegion(d)
		},
	}
}

func NewServer(d *schema.ResourceData) schema.CoreResource {
	return &sakura.Server{
		Address:    d.Address,
		Zone:       d.Get("zone").String(),
		Core:       d.Get("core").Int(),
		Memory:     d.Get("memory").Int(),
		Commitment: d.Get("commitment").String(),
		CPUModel:   d.Get("cpu_model").String(),
		GPUCount:   d.Get("gpu").Int(),
		GPUModel:   d.Get("gpu_model").String(),
	}
}
