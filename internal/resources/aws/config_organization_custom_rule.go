package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type ConfigOrganizationCustomRuleItem struct {
	Address                *string
	Region                 *string
	MonthlyRuleEvaluations *int64 `infracost_usage:"monthly_rule_evaluations"`
}

var ConfigOrganizationCustomRuleItemUsageSchema = []*schema.UsageItem{{Key: "monthly_rule_evaluations", ValueType: schema.Int64, DefaultValue: 0}}

func (r *ConfigOrganizationCustomRuleItem) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ConfigOrganizationCustomRuleItem) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := []*schema.CostComponent{}

	if r != nil && r.MonthlyRuleEvaluations != nil {
		monthlyConfigRules := decimal.NewFromInt(*r.MonthlyRuleEvaluations)

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
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: ConfigOrganizationCustomRuleItemUsageSchema,
	}
}
