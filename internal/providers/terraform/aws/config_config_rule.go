package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetConfigRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_config_rule",
		RFunc: NewConfigRule,
	}
}

func NewConfigRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_config_items").Exists() {
		monthlyConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_config_items").Int()))
	}

	var monthlyCustomConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_custom_config_items").Exists() {
		monthlyCustomConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_custom_config_items").Int()))
	}

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Config items",
		Unit:            "records",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ConfigurationItemRecorded/")},
			},
		},
	})

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Custom config items",
		Unit:            "records",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyCustomConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/CustomConfigItemRecorded/")},
			},
		},
	})

	if u != nil && u.Get("monthly_rule_evaluations").Exists() && u.Get("monthly_conformance_pack_evaluations").Exists() {
		monthlyConfigRules := decimal.NewFromInt(u.Get("monthly_rule_evaluations").Int())
		monthlyConformancePacks := decimal.NewFromInt(u.Get("monthly_conformance_pack_evaluations").Int())

		configRulesLimits := []int{100000, 400000}
		conformancePacksLimits := []int{1000000, 24000000}

		rulesTiers := usage.CalculateTierBuckets(monthlyConfigRules, configRulesLimits)
		packsTiers := usage.CalculateTierBuckets(monthlyConformancePacks, conformancePacksLimits)

		if rulesTiers[0].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Config rule evaluations (first 100K)", "0", &rulesTiers[0]))
		}
		if rulesTiers[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Config rule evaluations (next 400K)", "100000", &rulesTiers[1]))
		}
		if rulesTiers[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configRulesCostComponent(region, "Config rule evaluations (over 500K)", "500000", &rulesTiers[2]))
		}
		if packsTiers[0].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configPacksCostComponent(region, "Conformance pack evaluations (first 1M)", "0", &packsTiers[0]))
		}
		if packsTiers[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configPacksCostComponent(region, "Conformance pack evaluations (next 24M)", "1000000", &packsTiers[1]))
		}
		if packsTiers[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, configPacksCostComponent(region, "Conformance pack evaluations (over 25M)", "25000000", &packsTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, configRulesCostComponent(region, "Config rule evaluations (first 100K)", "0", unknown))
		costComponents = append(costComponents, configPacksCostComponent(region, "Conformance pack evaluations (first 1M)", "0", unknown))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func configRulesCostComponent(region string, displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "evaluations",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config Rules"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ConfigRuleEvaluations/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}

func configPacksCostComponent(region string, displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "evaluations",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config Conformance Packs"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ConformancePackEvaluations/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
