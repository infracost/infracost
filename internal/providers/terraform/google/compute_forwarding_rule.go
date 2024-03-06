package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeForwardingRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_forwarding_rule",
		CoreRFunc: NewComputeForwardingRule,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeGlobalForwardingRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_global_forwarding_rule",
		CoreRFunc: NewComputeForwardingRule,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}

func NewComputeForwardingRule(d *schema.ResourceData) schema.CoreResource {
	r := &google.ComputeForwardingRule{Address: d.Address, Region: d.Get("region").String()}
	return r
}
