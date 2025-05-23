package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGrafanaWorkspaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_grafana_workspace",
		CoreRFunc: NewGrafanaWorkspace,
	}
}

func NewGrafanaWorkspace(d *schema.ResourceData) schema.CoreResource {
	r := &aws.GrafanaWorkspace{
		Address: d.Address,
		Region:  d.Get("region").String(),
		License: d.Get("license_type").String(),
	}

	// Set default license type if not specified
	if r.License == "" {
		r.License = "ENTERPRISE" // Default to ENTERPRISE as per AWS documentation
	}

	return r
} 