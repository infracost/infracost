package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMLoadBalancerOutboundRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb_outbound_rule",
		RFunc: NewAzureRMLoadBalancerRule,
		ReferenceAttributes: []string{
			"loadbalancer_id",
			"resource_group_name",
		},
	}
}

func NewAzureRMLoadBalancerOutboundRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"loadbalancer_id", "resource_group_name"})
	region = convertRegion(region)

	var costComponents []*schema.CostComponent
	costComponents = append(costComponents, rulesCostComponent(region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
