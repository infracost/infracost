package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Wafv2WebACL struct {
	Address                                          *string
	Region                                           *string
	Rule0ActionLen                                   *int64
	Rule0Statement0RuleGroupReferenceStatement       *string
	RuleGroupReferenceStatementsCount                *int64
	Rule0Statement0ManagedRuleGroupStatement0NameLen *int64
	ManagedRuleGroupRules                            *int64 `infracost_usage:"managed_rule_group_rules"`
	MonthlyRequests                                  *int64 `infracost_usage:"monthly_requests"`
	RuleGroupRules                                   *int64 `infracost_usage:"rule_group_rules"`
}

var Wafv2WebACLUsageSchema = []*schema.UsageItem{{Key: "managed_rule_group_rules", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "rule_group_rules", ValueType: schema.Int64, DefaultValue: 0}}

func (r *Wafv2WebACL) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Wafv2WebACL) BuildResource() *schema.Resource {
	region := *r.Region

	var costComponents []*schema.CostComponent
	var ruleGroupRules, managedRuleGroupRules, monthlyRequests, rule *decimal.Decimal
	var sumForRules decimal.Decimal

	costComponents = append(costComponents, WafWebACLUsageCostComponent(
		region,
		"Web ACL usage",
		"months",
		"[A-Z0-9]*-(?!ShieldProtected-)WebACLV2",
		1,
		decimalPtr(decimal.NewFromInt(1)),
	))
	if r.Rule0ActionLen != nil {
		rule = decimalPtr(decimal.NewFromInt(*r.Rule0ActionLen))
	}
	if r != nil && r.RuleGroupRules != nil && r.ManagedRuleGroupRules != nil {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(*r.RuleGroupRules))
		managedRuleGroupRules = decimalPtr(decimal.NewFromInt(*r.ManagedRuleGroupRules))
		sumForRules = ruleGroupRules.Add(*managedRuleGroupRules)
		if rule.IsPositive() {
			sumForRules = sumForRules.Add(*rule)
		}
	}

	if sumForRules.IsPositive() {
		costComponents = append(costComponents, WafWebACLUsageCostComponent(
			region,
			"Rules",
			"rules",
			"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
			1,
			&sumForRules,
		))
	}

	if r.Rule0Statement0RuleGroupReferenceStatement != nil {
		counter := *r.RuleGroupReferenceStatementsCount

		if counter > 0 {
			costComponents = append(costComponents, WafWebACLUsageCostComponent(
				region,
				"Rule groups",
				"groups",
				"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
				1,
				decimalPtr(decimal.NewFromInt(int64(counter))),
			))
		}
	}

	if *r.Rule0Statement0ManagedRuleGroupStatement0NameLen > 0 {
		costComponents = append(costComponents, WafWebACLUsageCostComponent(
			region,
			"Managed rule groups",
			"groups",
			"[A-Z0-9]*-(?!ShieldProtected-)RuleV2",
			1,
			decimalPtr(decimal.NewFromInt(*r.Rule0Statement0ManagedRuleGroupStatement0NameLen)),
		))
	}

	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	costComponents = append(costComponents, WafWebACLUsageCostComponent(
		region,
		"Requests",
		"1M requests",
		"[A-Z0-9]*-(?!ShieldProtected-)RequestV2-Tier1",
		1000000,
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: Wafv2WebACLUsageSchema,
	}
}
