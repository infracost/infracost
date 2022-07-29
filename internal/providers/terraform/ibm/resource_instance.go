package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getResourceInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_resource_instance",
		RFunc: newResourceInstance,
	}
}

func newResourceInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &ibm.ResourceInstance{
		Address:    d.Address,
		Service:    d.Get("service").String(),
		Plan:       d.Get("plan").String(),
		Location:   d.Get("location").String(),
		Parameters: d.RawValues,
	}
	r.PopulateUsage(u)
	SetCatalogMetadata(d, d.Get("service").String())

	return r.BuildResource()
}
