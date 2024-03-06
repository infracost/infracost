package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getLoggingOrganizationSinkRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_logging_organization_sink",
		CoreRFunc: NewLoggingOrganizationSink,
	}
}

func NewLoggingOrganizationSink(d *schema.ResourceData) schema.CoreResource {
	r := &google.Logging{
		Address: d.Address,
	}

	return r
}
