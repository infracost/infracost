package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGrafanaWorkspaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_grafana_workspace",
		CoreRFunc:           NewGrafanaWorkspace,
		ReferenceAttributes: []string{"aws_grafana_license_association.workspace_id"},
	}
}

func NewGrafanaWorkspace(d *schema.ResourceData) schema.CoreResource {
	licenseType := "STANDARD"
	licenseAssoc := d.References("aws_grafana_license_association.workspace_id")
	if len(licenseAssoc) > 0 {
		licenseType = licenseAssoc[0].Get("license_type").String()
	}

	r := &aws.GrafanaWorkspace{
		Address: d.Address,
		Region:  d.Get("region").String(),
		License: licenseType,
	}

	return r
}
