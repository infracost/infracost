package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDnsResolverInboundEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_private_dns_resolver_inbound_endpoint",
		CoreRFunc: newPrivateDnsResolverInboundEndpoint,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newPrivateDnsResolverInboundEndpoint(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	} else if strings.HasPrefix(strings.ToLower(region), "germany") {
		region = "DE Zone 1"
	} else if strings.HasPrefix(strings.ToLower(region), "china") {
		region = "Zone 1 (China)"
	} else {
		region = "Zone 1"
	}

	return &azure.PrivateDnsResolverInboundEndpoint{
		Address: d.Address,
		Region:  region,
	}
}
