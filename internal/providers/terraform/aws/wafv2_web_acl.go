package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getWAFv2WebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_wafv2_web_acl",
		CoreRFunc: NewWAFv2WebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWAFv2WebACL(d *schema.ResourceData) schema.CoreResource {
	rules := int64(0)
	ruleGroups := int64(0)
	managedRuleGroups := int64(0)

	for _, rule := range d.Get("rule").Array() {
		if len(rule.Get("statement.0.rule_group_reference_statement").Array()) > 0 {
			ruleGroups++
		}

		if len(rule.Get("statement.0.managed_rule_group_statement").Array()) > 0 {
			managedRuleGroups++
		}

		// If the rule is neither a rule group or a managed rule group, it is a rule.
		if len(rule.Get("statement.0.rule_group_reference_statement").Array()) == 0 && len(rule.Get("statement.0.managed_rule_group_statement").Array()) == 0 {
			rules++
		}
	}

	r := &aws.WAFv2WebACL{
		Address:           d.Address,
		Region:            d.Get("region").String(),
		Rules:             rules,
		RuleGroups:        ruleGroups,
		ManagedRuleGroups: managedRuleGroups,
	}
	return r
}
