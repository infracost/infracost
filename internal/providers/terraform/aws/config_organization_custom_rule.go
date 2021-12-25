package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetConfigOrganizationCustomRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_organization_custom_rule",
		RFunc: NewConfigOrganizationCustomRuleItem,
	}
}
func NewConfigOrganizationCustomRuleItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ConfigOrganizationCustomRuleItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
