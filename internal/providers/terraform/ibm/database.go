package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_database",
		RFunc: newDatabase,
	}
}

func newDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	plan := d.Get("plan").String()
	location := d.Get("location").String()
	service := d.Get("service").String()
	name := d.Get("name").String()

	r := &ibm.Database{
		Name:     name,
		Address:  d.Address,
		Service:  service,
		Plan:     plan,
		Location: location,
		Group:    d.RawValues,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["service"] = service
	configuration["plan"] = plan
	configuration["location"] = location

	SetCatalogMetadata(d, service, configuration)

	return r.BuildResource()
}
