package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetWafv2WebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_wafv2_web_acl",
		RFunc: NewWafv2WebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWafv2WebACL(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var costComponents []*schema.CostComponent
	var ruleGroupRules, managedRuleGroupRules, monthlyRequests, rule *decimal.Decimal
	var sumForRules decimal.Decimal

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Web ACL usage",
		"months",
		"[A-Z0-9]*-(?!ShieldProtected-)WebACLV2",
		1,
		decimalPtr(decimal.NewFromInt(1)),
	))
	if d.Get("rule.0.action").Type != gjson.Null {
		rules := d.Get("rule.0.action").Array()
		rule = decimalPtr(decimal.NewFromInt(int64(len(rules))))
	}
	if u != nil && u.Get("rule_group_rules").Type != gjson.Null && u.Get("managed_rule_group_rules").Type != gjson.Null {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("rule_group_rules").Int()))
		managedRuleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("managed_rule_group_rules").Int()))
		sumForRules = ruleGroupRules.Add(*managedRuleGroupRules)
		if rule.IsPositive() {
			sumForRules = sumForRules.Add(*rule)
		}
	}

	if sumForRules.IsPositive() {
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Rules",
			"rules",
			"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
			1,
			&sumForRules,
		))
	}

	if d.Get("rule.0.statement.0.rule_group_reference_statement").Type != gjson.Null {
		counter := 0
		if d.Get("rule").Type != gjson.Null {
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
		}

		if counter > 0 {
			costComponents = append(costComponents, wafWebACLUsageCostComponent(
				region,
				"Rule groups",
				"groups",
				"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
				1,
				decimalPtr(decimal.NewFromInt(int64(counter))),
			))
		}
	}
	manageQuantity := d.Get("rule.0.statement.0.managed_rule_group_statement.0.name").Array()

	if len(manageQuantity) > 0 {
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Managed rule groups",
			"groups",
			"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
			1,
			decimalPtr(decimal.NewFromInt(int64(len(manageQuantity)))),
		))
	}

	if u != nil && u.Get("monthly_requests").Type != gjson.Null {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Requests",
		"1M requests",
		"[A-Z0-9]*-(?!ShieldProtected-)RequestV2-Tier1",
		1000000,
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
