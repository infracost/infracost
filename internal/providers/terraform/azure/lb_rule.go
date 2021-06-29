package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMLoadBalancerRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb_rule",
		RFunc: NewAzureRMLoadBalancerRule,
		ReferenceAttributes: []string{
			"loadbalancer_id",
			"resource_group_name",
		},
	}
}

func NewAzureRMLoadBalancerRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"loadbalancer_id", "resource_group_name"})
	if len(region) > 0 {
		if strings.HasPrefix(strings.ToLower(region), "usgov") {
			region = "US Gov"
		} else if strings.Contains(strings.ToLower(region), "china") {
			region = "Сhina"
		} else {
			region = "Global"
		}
	}

	var costComponents []*schema.CostComponent
	costComponents = append(costComponents, rulesCostComponent(region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func rulesCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Rule usage",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Load Balancer"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Overage LB Rules and Outbound Rules")},
			},
		},
	}
}
