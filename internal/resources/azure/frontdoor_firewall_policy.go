package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// FrontdoorFirewallPolicy represents a policy for Web Application Firewall (WAF)
// with Azure Front Door.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/web-application-firewall/afds/waf-front-door-drs
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/frontdoor/#overview
type FrontdoorFirewallPolicy struct {
	Address string
	Region  string

	CustomRules     int
	ManagedRulesets int

	// "usage" args
	MonthlyCustomRuleRequests     *int64 `infracost_usage:"monthly_custom_rule_requests"`
	MonthlyManagedRulesetRequests *int64 `infracost_usage:"monthly_managed_ruleset_requests"`
}

// CoreType returns the name of this resource type
func (r *FrontdoorFirewallPolicy) CoreType() string {
	return "FrontdoorFirewallPolicy"
}

// UsageSchema defines a list which represents the usage schema of EventGridTopic.
func (r *FrontdoorFirewallPolicy) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_custom_rule_requests", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_managed_ruleset_requests", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the FrontdoorFirewallPolicy.
func (r *FrontdoorFirewallPolicy) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid FrontdoorFirewallPolicy.
// This method is called after the resource is initialised by an IaC provider.
func (r *FrontdoorFirewallPolicy) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.policyCostComponents()...)
	costComponents = append(costComponents, r.customRulesCostComponents()...)
	costComponents = append(costComponents, r.customRuleRequestsCostComponents()...)
	costComponents = append(costComponents, r.managedRulesetsCostComponents()...)
	costComponents = append(costComponents, r.managedRulesetRequestsCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// policyCostComponents returns cost components for Policy usage
func (r *FrontdoorFirewallPolicy) policyCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Policy",
			Unit:            "months",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter:   r.buildProductFilter("Policy"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}
}

// customRulesCostComponents returns a cost component for the total number of custom
// rules in the policy.
func (r *FrontdoorFirewallPolicy) customRulesCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Custom rules",
			Unit:            "rules",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(r.CustomRules))),
			ProductFilter:   r.buildProductFilter("Rule"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}
}

// customRuleRequestsCostComponents returns a usage based cost component for the
// number of custom rules' requests.
func (r *FrontdoorFirewallPolicy) customRuleRequestsCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Custom rule requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.monthlyRequestsQuantity(r.MonthlyCustomRuleRequests),
			ProductFilter:   r.buildProductFilter("Requests"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
			UsageBased: true,
		},
	}
}

// managedRulesetsCostComponents returns a cost component for the total number
// of managed rulesets in the policy.
func (r *FrontdoorFirewallPolicy) managedRulesetsCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Managed rulesets",
			Unit:            "rulesets",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(r.ManagedRulesets))),
			ProductFilter:   r.buildProductFilter("Default Ruleset"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}
}

// managedRulesetRequestsCostComponents returns a usage based cost component for
// the number of managed rulesets' requests.
func (r *FrontdoorFirewallPolicy) managedRulesetRequestsCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Managed ruleset requests",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: r.monthlyRequestsQuantity(r.MonthlyManagedRulesetRequests),
			ProductFilter:   r.buildProductFilter("Default Request"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
			UsageBased: true,
		},
	}
}

// buildProductFilter returns a product filter for the Front Door's products.
//
// skuName and productName define the original Front Door service (not
// Standard/Premium).
func (r *FrontdoorFirewallPolicy) buildProductFilter(meterName string) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:    strPtr("azure"),
		Region:        strPtr(r.Region),
		Service:       strPtr("Azure Front Door Service"),
		ProductFamily: strPtr("Networking"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "skuName", Value: strPtr("Standard")},
			{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
			{Key: "productName", Value: strPtr("Azure Front Door Service")},
		},
	}
}

// monthlyRequestsQuantity converts the monthly requests usage number as
// Azure's requests pricing is 1M requests/month.
func (r *FrontdoorFirewallPolicy) monthlyRequestsQuantity(requestsNumber *int64) *decimal.Decimal {
	var monthlyRequests *decimal.Decimal
	divider := decimal.NewFromInt(1000000)

	if requestsNumber != nil {
		requests := decimal.NewFromInt(*requestsNumber)
		monthlyRequests = decimalPtr(requests.Div(divider))
	}

	return monthlyRequests
}
