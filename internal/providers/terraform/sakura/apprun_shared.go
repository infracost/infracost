package sakura

import (
	"github.com/infracost/infracost/internal/resources/sakura"
	"github.com/infracost/infracost/internal/schema"
)

func getApprunSharedRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "sakura_apprun_shared",
		CoreRFunc: NewApprunShared,
	}
}

func NewApprunShared(d *schema.ResourceData) schema.CoreResource {
	components := parseAppComponents(d)
	return &sakura.ApprunShared{
		Address:    d.Address,
		MinScale:   d.Get("min_scale").Int(),
		MaxScale:   d.Get("max_scale").Int(),
		Components: components,
	}
}

func parseAppComponents(d *schema.ResourceData) []sakura.AppComponent {
	raw := d.Get("components").Array()
	components := make([]sakura.AppComponent, 0, len(raw))
	for _, item := range raw {
		components = append(components, sakura.AppComponent{
			MaxCPU:    item.Get("max_cpu").String(),
			MaxMemory: item.Get("max_memory").String(),
		})
	}
	return components
}
