package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSRecordSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_dns_record_set",
		CoreRFunc: NewDNSRecordSet,
	}
}

func NewDNSRecordSet(d *schema.ResourceData) schema.CoreResource {
	r := &google.DNSRecordSet{
		Address: d.Address,
	}

	return r
}
