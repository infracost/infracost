package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getWafv2WebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_wafv2_web_acl",
		RFunc: NewWafv2WebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}
func NewWafv2WebACL(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Wafv2WebACL{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), Rule0Statement0ManagedRuleGroupStatement0NameLen: intPtr(int64(len(d.Get("rule.0.statement.0.managed_rule_group_statement.0.name").Array())))}
	if !d.IsEmpty("rule.0.action") {
		r.Rule0ActionLen = intPtr(int64(len(d.Get("rule.0.action").Array())))
	}
	if !d.IsEmpty("rule.0.statement.0.rule_group_reference_statement") {
		r.Rule0Statement0RuleGroupReferenceStatement = strPtr(d.Get("rule.0.statement.0.rule_group_reference_statement").String())
		var counter int64
		rules := d.Get("rule").Array()
		for _, rule := range rules {
			if rule.Get("statement").Type != gjson.Null {
				statements := rule.Get("statement").Array()
				for _, statement := range statements {
					if statement.Get("rule_group_reference_statement").Type != gjson.Null {
						counter++
					}
				}
			}
		}
		r.RuleGroupReferenceStatementsCount = &counter
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
