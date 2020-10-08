package aws

import "github.com/infracost/infracost/pkg/schema"

var (
	freeResourcesList []string = []string{}
)

func GetFreeResources() []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range freeResourcesList {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:   resourceName,
			NoCost: true,
		})
	}
	return freeResources
}
