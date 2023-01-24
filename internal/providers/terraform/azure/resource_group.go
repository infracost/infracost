package azure

import "github.com/infracost/infracost/internal/schema"

// getResourceGroupDefinitionRegistryItem defines a free resource which is required to create a custo ReferenceAttributes via the name.
func getResourceGroupDefinitionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:    "azurerm_resource_group",
		NoPrice: true,
		Notes:   []string{"Free resource."},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			var refs []string
			name := d.Get("name").String()
			if name != "" {
				refs = append(refs, name)
			}

			id := d.Get("id").String()
			if id != "" {
				refs = append(refs, id)
			}

			return refs
		},
	}
}
