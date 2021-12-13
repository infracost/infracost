package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetConfigOrganizationCustomRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_organization_custom_rule",
		RFunc: NewConfigOrganizationCustomRule,
	}
}

func NewConfigOrganizationCustomRule(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{}

	if u != nil && u.Get("monthly_rule_evaluations").Exists() {
		monthlyConfigRules := decimal.NewFromInt(u.Get("monthly_rule_evaluations").Int())

		configRulesLimits := []int{100000, 400000}

		rulesTiers := usage.CalculateTierBuckets(monthlyConfigRules, configRulesLimits)

		if rulesTiers[0].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Rule evaluations (first 100K)", "0", &rulesTiers[0]))
		}
		if rulesTiers[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Rule evaluations (next 400K)", "100000", &rulesTiers[1]))
		}
		if rulesTiers[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Rule evaluations (over 500K)", "500000", &rulesTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, configRulesCostComponent(region, "Rule evaluations (first 100K)", "0", unknown))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
