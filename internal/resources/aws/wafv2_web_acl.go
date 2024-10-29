package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type WAFv2WebACL struct {
	Address               string
	Region                string
	Rules                 int64
	RuleGroups            int64
	ManagedRuleGroups     int64
	RuleGroupRules        *int64 `infracost_usage:"rule_group_rules"`
	ManagedRuleGroupRules *int64 `infracost_usage:"managed_rule_group_rules"`
	MonthlyRequests       *int64 `infracost_usage:"monthly_requests"`
}

func (r *WAFv2WebACL) CoreType() string {
	return "WAFv2WebACL"
}

func (r *WAFv2WebACL) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "managed_rule_group_rules", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "rule_group_rules", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *WAFv2WebACL) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WAFv2WebACL) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{r.webACLUsageCostComponent()}

	rules := r.Rules
	if r.RuleGroupRules != nil {
		rules += *r.RuleGroupRules
	}
	if r.ManagedRuleGroupRules != nil {
		rules += *r.ManagedRuleGroupRules
	}

	if rules > 0 {
		costComponents = append(costComponents, r.rulesCostComponent(rules))
	}

	if r.RuleGroups > 0 {
		costComponents = append(costComponents, r.ruleGroupsCostComponent("Rule groups", r.RuleGroups))
	}

	if r.ManagedRuleGroups > 0 {
		costComponents = append(costComponents, r.ruleGroupsCostComponent("Managed rule groups", r.RuleGroups))
	}

	costComponents = append(costComponents, r.requestsCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *WAFv2WebACL) webACLUsageCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Web ACL usage",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(int64(1)),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)WebACLV2$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *WAFv2WebACL) rulesCostComponent(rules int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Rules",
		Unit:            "rules",
		UnitMultiplier:  decimal.NewFromInt(int64(1)),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(rules)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)RuleV2$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *WAFv2WebACL) ruleGroupsCostComponent(name string, ruleGroups int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "groups",
		UnitMultiplier:  decimal.NewFromInt(int64(1)),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(ruleGroups)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)RuleV2$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *WAFv2WebACL) requestsCostComponent() *schema.CostComponent {
	var requests *decimal.Decimal
	if r.MonthlyRequests != nil {
		requests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(int64(1000000)),
		MonthlyQuantity: requests,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)RequestV2-Tier1$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
