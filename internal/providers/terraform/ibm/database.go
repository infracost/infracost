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

	var members int64
	if service == "databases-for-elasticsearch" {
		members = 3
	} else if service == "databases-for-postgresql" {
		members = 2
	}

	var flavor string
	var memory int64
	var cpu int64
	var disk int64

	for _, g := range d.Get("group").Array() {

		if g.Get("group_id").String() == "member" {

			if len(g.Get("host_flavor").Array()) > 0 {
				flavor = g.Get("host_flavor").Array()[0].Map()["id"].String()
			}
			if len(g.Get("memory").Array()) > 0 {
				memory = g.Get("memory").Array()[0].Map()["allocation_mb"].Int()
			}
			if len(g.Get("cpu").Array()) > 0 {
				cpu = g.Get("cpu").Array()[0].Map()["allocation_count"].Int()
			}
			if len(g.Get("members").Array()) > 0 {
				members = g.Get("members").Array()[0].Map()["allocation_count"].Int()
			}
			if len(g.Get("disk").Array()) > 0 {
				disk = g.Get("disk").Array()[0].Map()["allocation_mb"].Int()
			}
		}
	}

	r := &ibm.Database{
		Name:     name,
		Address:  d.Address,
		Service:  service,
		Plan:     plan,
		Location: location,
		Group:    d.RawValues,
		Flavor:   flavor,
		Disk:     disk,
		Memory:   memory,
		CPU:      cpu,
		Members:  members,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["service"] = service
	configuration["plan"] = plan
	configuration["location"] = location
	configuration["disk"] = disk
	configuration["members"] = members

	if flavor != "" {
		configuration["flavor"] = flavor
	} else {
		configuration["memory"] = memory
		configuration["cpu"] = cpu
	}

	SetCatalogMetadata(d, service, configuration)

	return r.BuildResource()
}
