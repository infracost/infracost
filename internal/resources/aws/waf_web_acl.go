package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type WAFWebACL struct {
	Address         string
	Region          string
	Rules           int64
	RuleGroups      int64
	RuleGroupRules  *int64 `infracost_usage:"rule_group_rules"`
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

func (r *WAFWebACL) CoreType() string {
	return "WAFWebACL"
}

func (r *WAFWebACL) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "rule_group_rules", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *WAFWebACL) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WAFWebACL) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{r.webACLUsageCostComponent()}

	rules := r.Rules
	if r.RuleGroupRules != nil {
		rules += *r.RuleGroupRules
	}

	costComponents = append(costComponents, r.rulesCostComponent(rules))
	costComponents = append(costComponents, r.ruleGroupsCostComponent(r.RuleGroups))
	costComponents = append(costComponents, r.requestsCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *WAFWebACL) webACLUsageCostComponent() *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)WebACL$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *WAFWebACL) rulesCostComponent(rules int64) *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)Rule$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *WAFWebACL) ruleGroupsCostComponent(ruleGroups int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Rule groups",
		Unit:            "groups",
		UnitMultiplier:  decimal.NewFromInt(int64(1)),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(ruleGroups)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)Rule$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *WAFWebACL) requestsCostComponent() *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: strPtr("/^[A-Z0-9]*-(?!ShieldProtected-)Request$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
