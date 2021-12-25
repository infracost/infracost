package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetConfigOrganizationManagedRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_organization_managed_rule",
		RFunc: NewConfigOrganizationManagedRuleItem,
	}
}
func NewConfigOrganizationManagedRuleItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ConfigOrganizationManagedRuleItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
