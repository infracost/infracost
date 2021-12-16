package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getArtifactRegistryRepositoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_artifact_registry_repository",
		RFunc: newArtifactRegistryRepository,
	}
}

func newArtifactRegistryRepository(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}

	location := d.Get("location").String()
	if location != "" {
		region = location
	}

	r := &google.ArtifactRegistryRepository{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
