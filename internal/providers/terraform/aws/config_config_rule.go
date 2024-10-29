package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getConfigRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_config_config_rule",
		CoreRFunc: NewConfigConfigRule,
	}
}
func NewConfigConfigRule(d *schema.ResourceData) schema.CoreResource {
	r := &aws.ConfigConfigRule{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
