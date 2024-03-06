package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getConfigOrganizationManagedRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_config_organization_managed_rule",
		CoreRFunc: NewConfigOrganizationManagedRule,
	}
}
func NewConfigOrganizationManagedRule(d *schema.ResourceData) schema.CoreResource {
	r := &aws.ConfigConfigRule{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
