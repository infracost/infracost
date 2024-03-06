package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type ConfigConfigRule struct {
	Address                string
	Region                 string
	MonthlyRuleEvaluations *int64 `infracost_usage:"monthly_rule_evaluations"`
}

func (r *ConfigConfigRule) CoreType() string {
	return "ConfigConfigRule"
}

func (r *ConfigConfigRule) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_rule_evaluations", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *ConfigConfigRule) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ConfigConfigRule) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if r.MonthlyRuleEvaluations != nil {
		monthlyConfigRules := decimal.NewFromInt(*r.MonthlyRuleEvaluations)

		configRulesLimits := []int{100000, 400000}

		rulesTiers := usage.CalculateTierBuckets(monthlyConfigRules, configRulesLimits)

		if rulesTiers[0].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.configRulesCostComponent("Rule evaluations (first 100K)", "0", &rulesTiers[0]))
		}
		if rulesTiers[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.configRulesCostComponent("Rule evaluations (next 400K)", "100000", &rulesTiers[1]))
		}
		if rulesTiers[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.configRulesCostComponent("Rule evaluations (over 500K)", "500000", &rulesTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, r.configRulesCostComponent("Rule evaluations (first 100K)", "0", unknown))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ConfigConfigRule) configRulesCostComponent(displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "evaluations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config Rules"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-ConfigRuleEvaluations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
