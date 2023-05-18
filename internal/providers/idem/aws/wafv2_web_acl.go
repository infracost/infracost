package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetWAFv2WebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.wafv2.web_acl.present",
		RFunc: NewWAFv2WebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Idem.",
		},
	}
}

func NewWAFv2WebACL(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	rules := int64(0)
	ruleGroups := int64(0)
	managedRuleGroups := int64(0)

	for _, rule := range d.Get("rules").Array() {
		if len(rule.Get("Statement.RuleGroupReferenceStatement").Array()) > 0 {
			ruleGroups++
		}

		if len(rule.Get("Statement.ManagedRuleGroupStatement").Array()) > 0 {
			managedRuleGroups++
		}

		// If the rule is neither a rule group or a managed rule group, it is a rule.
		if len(rule.Get("Statement.RuleGroupReferenceStatement").Array()) == 0 && len(rule.Get("Statement.ManagedRuleGroupStatement").Array()) == 0 {
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

	r.PopulateUsage(u)
	return r.BuildResource()
}
