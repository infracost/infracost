package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetConfigRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.config.rule.present",
		RFunc: NewConfigConfigRule,
	}
}
func NewConfigConfigRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ConfigConfigRule{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
