package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIsShareRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_share",
		RFunc: newIsShare,
		ReferenceAttributes: []string{
			"source_share",
			"source_share.0.size",
		}}
}

func newIsShare(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	size := d.Get("size").Int()
	profile := d.Get("profile").String()
	iops := d.Get("iops").Int()
	zone := d.Get("zone").String()
	name := d.Get("name").String()

	// if the reference to a source share can
	// be resolved, then the share is a replica and the size is specified by the source
	sourceShareRef := d.References("source_share")
	isReplicaShare := len(sourceShareRef) > 0
	if isReplicaShare {
		sourceShare := sourceShareRef[0]
		size = sourceShare.Get("size").Int()
	}

	// if an inline replica share can be found, then
	// estimate for 2 shares @TODO get the zone info from the replica_zone section in the TF
	var replicaZone = ""
	var replicaName = ""
	if !d.IsEmpty("replica_share") {
		replicaZone = d.Get("replica_share.0.zone").String()
		replicaName = d.Get("replica_share.0.name").String()
	}

	r := &ibm.IsShare{
		Address:           d.Address,
		Region:            region,
		Profile:           profile,
		IOPS:              iops,
		Size:              size,
		Zone:              zone,
		IsReplica:         isReplicaShare,
		InlineReplicaZone: replicaZone,
		InlineReplicaName: replicaName,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["profile"] = profile
	configuration["size"] = size
	configuration["iops"] = iops
	configuration["zone"] = zone
	configuration["name"] = name
	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
