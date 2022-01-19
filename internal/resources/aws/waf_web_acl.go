package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type WafWebAcl struct {
	Address         *string
	Region          *string
	RulesTypes      *[]string
	RuleGroupRules  *int64 `infracost_usage:"rule_group_rules"`
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

var WafWebAclUsageSchema = []*schema.UsageItem{{Key: "rule_group_rules", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}}

func (r *WafWebAcl) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WafWebAcl) BuildResource() *schema.Resource {
	region := *r.Region

	var costComponents []*schema.CostComponent
	var ruleGroupRules, monthlyRequests, rule *decimal.Decimal

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Web ACL usage",
		"months",
		"[A-Z0-9]*-(?!ShieldProtected-)WebACL",
		1,
		decimalPtr(decimal.NewFromInt(1)),
	))
	if r.RulesTypes != nil {
		var count int
		for _, types := range *r.RulesTypes {
			if strings.ToLower(types) == "regular" || strings.ToLower(types) == "rate_based" {
				count++
			}

		}
		rule = decimalPtr(decimal.NewFromInt(int64(count)))
	}

	if r.RuleGroupRules != nil {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(*r.RuleGroupRules))
		rule = decimalPtr(ruleGroupRules.Add(*rule))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Rules",
		"rules",
		"[A-Z0-9]*-(?!ShieldProtected-)Rule",
		1,
		rule,
	))

	var count int
	if r.RulesTypes != nil {
		for _, types := range *r.RulesTypes {
			if strings.ToLower(types) == "group" {
				count++
			}

		}
	}

	if count > 0 {
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Rule groups",
			"groups",
			"[A-Z0-9]*-(?!ShieldProtected-)Rule",
			1,
			decimalPtr(decimal.NewFromInt(int64(count))),
		))
	}

	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Requests",
		"1M requests",
		"[A-Z0-9]*-(?!ShieldProtected-)Request",
		1000000,
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: WafWebAclUsageSchema,
	}
}

func wafWebACLUsageCostComponent(region, displayName, unit, usagetype string, unitMultiplier int, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(int64(unitMultiplier)),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", usagetype))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
