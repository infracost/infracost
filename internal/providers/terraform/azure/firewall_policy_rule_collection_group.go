package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMFirewallPolicyRuleCollectionGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "azurerm_firewall_policy_rule_collection_group",
		RFunc:               newAzureRMFirewallPolicyRuleCollectionGroup,
		ReferenceAttributes: []string{"firewall_policy_id"},
	}
}

func newAzureRMFirewallPolicyRuleCollectionGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		DefaultTags:  d.DefaultTags,
		IsSkipped:    true,
		NoPrice:      true,
		SkipMessage:  "Free resource.",
	}
}
