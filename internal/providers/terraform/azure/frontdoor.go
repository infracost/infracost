package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

// getAzureRMFrontdoorRegistryItem returns a registry item for the resource
func getAzureRMFrontdoorRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_frontdoor",
		RFunc: newFrontdoor,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

// newFrontdoor parses Terraform's data and uses it to build a new resource
func newFrontdoor(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	} else {
		region = regionToZone(region)
	}

	rulesCounter := 0
	rules := d.Get("routing_rule").Array()
	for _, rule := range rules {
		if rule.Get("enabled").Type == gjson.True {
			rulesCounter++
		}
	}

	r := &azure.Frontdoor{
		Address:       d.Address,
		Region:        region,
		FrontendHosts: len(d.Get("frontend_endpoint").Array()),
		RoutingRules:  rulesCounter,
	}
	r.PopulateUsage(u)

	return r.BuildResource(ctx)
}
