package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSManagedZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_dns_managed_zone",
		RFunc: NewDNSManagedZone,
	}
}

func NewDNSManagedZone(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.DNSManagedZone{
		Address: d.Address,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
