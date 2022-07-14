package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVirtualNetworkPeeringRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_network_peering",
		RFunc: newVirtualNetworkPeering,
		ReferenceAttributes: []string{
			"virtual_network_name",
			"remote_virtual_network_id",
			"resource_group_name",
		},
	}
}

func newVirtualNetworkPeering(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	sourceRegion := lookupRegion(d, []string{"virtual_network_name"})
	destinationRegion := lookupRegion(d, []string{"remote_virtual_network_id"})

	sourceZone := regionToZone(sourceRegion)
	destinationZone := regionToZone(destinationRegion)

	r := &azure.VirtualNetworkPeering{
		Address:           d.Address,
		DestinationRegion: destinationRegion,
		SourceRegion:      sourceRegion,
		DestinationZone:   destinationZone,
		SourceZone:        sourceZone,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
