package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIsLbRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_lb",
		RFunc: newIsLb,
	}
}

func newIsLb(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	var profile string
	profileStr := d.Get("profile").String()
	if profileStr == "network-fixed" {
		profile = "network"
	} else {
		profile = "application"
	}
	r := &ibm.IsLb{
		Address: d.Address,
		Region:  region,
		Profile: profile,
		Logging: d.GetBoolOrDefault("logging", false),
		Type:    d.Get("type").String(),
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["profile"] = profile

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
