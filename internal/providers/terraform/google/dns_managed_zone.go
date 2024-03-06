package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSManagedZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_dns_managed_zone",
		CoreRFunc: NewDNSManagedZone,
	}
}

func NewDNSManagedZone(d *schema.ResourceData) schema.CoreResource {
	r := &google.DNSManagedZone{
		Address: d.Address,
	}

	return r
}
